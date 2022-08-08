package vm

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/webappio/layerfiles/pkg/environment"
	"github.com/webappio/layerfiles/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type QemuVM struct {
	Cmd *exec.Cmd

	EnableKVM     bool
	CPU           string
	Memory        string
	NumProcessors int
	KernelFile    string

	commandHandler QEMUCommandHandler
	monitorHandler QEMUMonitorHandler
}

func (vm *QemuVM) GetHostIP() string {
	return "10.111.1.2"
}

func (vm *QemuVM) GetGuestIP() string {
	return "10.111.1.15" //TODO
}

func (vm *QemuVM) CreateQcowOverlay(base, target string) error {
	if base == target {
		return fmt.Errorf("base=target: %v", base)
	}
	_ = os.Remove(target)
	// https://events.static.linuxfound.org/sites/events/files/slides/kvm-forum-2017-slides.pdf
	out, err := exec.Command(
		"qemu-img", "create",
		"-o", "backing_file="+base+",backing_fmt=qcow2,lazy_refcounts=on,compat=1.1,cluster_size=128K",
		"-u",
		"-f", "qcow2",
		target,
		"60G",
	).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "error creating base image: %v", strings.TrimSpace(string(out)))
	}
	return nil
}

func SetupNetwork() error {
	interfaceName := "layerfile-net"

	err := util.RunCommandWithTimeout(exec.Command("sudo", "ip", "link", "del", interfaceName), time.Second)
	if err != nil && !strings.Contains(err.Error(), "Cannot find device") {
		return errors.Wrap(err, "could not delete interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo", "ip", "tuntap", "add", "dev", interfaceName, "mode", "tap"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not create network interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo", "ip", "addr", "add", "10.111.1.1/24", "dev", interfaceName), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not assign IP address to interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo", "ip", "link", "set", "dev", interfaceName, "up"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not start IP address for interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo",  "iptables", "-A", "FORWARD", "-p", "tcp", "-s", "10.111.1.1/30", "-j", "ACCEPT"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not use iptables to forward TCP traffic over "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo",  "iptables", "-A", "FORWARD", "-p", "udp", "-s", "10.111.1.1/30", "-j", "ACCEPT"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not use iptables to forward UDP traffic over  "+interfaceName)
	}

	//iptables -t nat -A PREROUTING -p tcp --dport 3124 -j DNAT --to-destination 1.1.1.1:3000
	err = util.RunCommandWithTimeout(exec.Command("sudo", "iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "DNAT", "--to-destination", "1.1.1.1:53"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not use iptables to forward DNS requests from "+interfaceName)
	}

	out, err := exec.Command("ip", "route", "get", "1.1.1.1").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "could not find internet interface: %v", strings.TrimSpace(string(out)))
	}
	inetInterface := strings.Fields(string(out))[4] //format is 1.1.1.1 via 192.168.86.1 dev wlp2s0 src 192.168.86.233 uid 1000

	err = util.RunCommandWithTimeout(exec.Command("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", inetInterface, "-j", "MASQUERADE"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not add a masquerade rule for traffic over wlp2s0")
	}

	return nil
}

func (vm *QemuVM) Start() error {
	diskLoc, err := environment.GetAndCreateDisksDirectory()
	if err != nil {
		return err
	}

	err = vm.CreateQcowOverlay(
		"/home/colin/projects/layerfiles/prepare-disks/ubuntu-22.04.qcow2", //TODO download the disk for the image
		filepath.Join(diskLoc, "disk.qcow2"),
		)
	if err != nil {
		return err
	}

	//err = SetupNetwork()
	//if err != nil {
	//	return err
	//}

	vm.Cmd = exec.Command("qemu-system-x86_64",
		"-M", "microvm,x-option-roms=off,isa-serial=off,rtc=off",
		"-no-acpi", //disable ACPI for faster boots
		"-enable-kvm", //use KVM for performance on linux
		"-cpu", "host", //faster CPU by reducing emulation
		"-nodefaults", //avoid default QEMU devices
		"-no-user-config", //do not read configuration files
		"-nographic", //do not display a window for the vm (background it)
		"-no-reboot",
		"-m", "512m", "-smp", "2",
		"-device", "virtio-serial-device",
		"-device", "virtio-rng-device", //add RNG from host to the vm
		"-chardev", "stdio,id=virtiocon0",
		"-device", "virtconsole,chardev=virtiocon0",
		"-kernel", "/home/colin/projects/linux-5.12.10/arch/x86_64/boot/bzImage",
		"-drive", "id=root,file="+filepath.Join(diskLoc, "disk.qcow2")+",format=qcow2,if=none",
		"-device", "virtio-blk-device,drive=root",
		"-append", "console=hvc0 root=/dev/vda rw acpi=off reboot=t panic=-1 ip=10.111.1.15::10.111.1.2:255.255.255.0:::off",
		//"-netdev", "tap,id=n0,ifname=layerfile-net,script=no,downscript=no",
		"-netdev", "user,id=n0,net=10.111.1.0/24,dhcpstart=10.111.1.15,hostfwd=tcp::30812-:30812",
		"-device", "virtio-net-device,netdev=n0,mac=52:54:00:12:34:56",
		"-monitor", "tcp:127.0.0.1:44531,server,nowait",
	)

	//for testing
	//vm.Cmd.Stdout = os.Stdout
	//vm.Cmd.Stderr = os.Stderr
	//vm.Cmd.Stdin = os.Stdin
	//vm.Cmd.Start()
	//time.Sleep(time.Hour)

	vm.commandHandler.Stdout, err = vm.Cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdout")
	}

	vm.Cmd.Stderr = os.Stderr

	vm.commandHandler.Stdin, err = vm.Cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdin")
	}

	err = vm.Cmd.Start()
	if err != nil {
		return errors.Wrap(err, "could not start VM")
	}

	var vmErr error
	vmDone := false
	go func() {
		vmErr = vm.Cmd.Wait()
		vmDone = true
	}()

	for i := 10; i >= 0; i -= 1 {
		err = vm.monitorHandler.Connect(44531)
		if err == nil {
			break
		}
		if i == 0 {
			return errors.Wrap(err, "could not connect to VM, did it start?")
		}
		if vmDone {
			return fmt.Errorf("vm never came up: %v", vmErr)
		}
		time.Sleep(time.Millisecond * 60)
	}

	return nil
}

func (vm *QemuVM) Stop() error {
	if vm.Cmd.Process == nil {
		return nil
	}
	return vm.Cmd.Process.Kill()
}

func (vm *QemuVM) GetCommandHandler() *QEMUCommandHandler {
	return &vm.commandHandler
}