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

	err = util.RunCommandWithTimeout(exec.Command("sudo", "ip", "addr", "add", "10.111.1.1/30", "dev", interfaceName), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not assign IP address to interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo", "ip", "link", "set", "dev", interfaceName, "up"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not start IP address for interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo",  "iptables", "-A", "FORWARD", "-p", "tcp", "-s", "10.111.1.1/30", "-j", "ACCEPT"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not start IP address for interface "+interfaceName)
	}

	err = util.RunCommandWithTimeout(exec.Command("sudo",  "iptables", "-A", "FORWARD", "-p", "udp", "-s", "10.111.1.1/30", "-j", "ACCEPT"), time.Second)
	if err != nil {
		return errors.Wrap(err, "could not start IP address for interface "+interfaceName)
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

	err = SetupNetwork()
	if err != nil {
		return err
	}

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
		"-append", "console=hvc0 root=/dev/vda rw acpi=off reboot=t panic=-1",
		"-netdev", "tap,id=tap0,ifname=layerfile-net,script=no,downscript=no",
		"-device", "virtio-net-device,netdev=tap0",
	)

	vm.commandHandler.Stdout, err = vm.Cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdout")
	}

	vm.Cmd.Stderr = os.Stderr

	vm.commandHandler.Stdin, err = vm.Cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdin")
	}

	return vm.Cmd.Start()
}

func (vm *QemuVM) Stop() error {
	if vm.Cmd.Process == nil {
		return nil
	}
	vm.Cmd.Process.Kill()
	_, err := vm.Cmd.Process.Wait()
	return err
}

func (vm *QemuVM) GetCommandHandler() *QEMUCommandHandler {
	return &vm.commandHandler
}

//func (vm *QemuVM) Save() error {
//
//}
//
//func (vm *QemuVM) Load(snap *snapshot.Snapshot) error {
//
//}
