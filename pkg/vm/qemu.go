package vm

import (
	"os"
	"os/exec"
)

type QemuVM struct {
	Cmd *exec.Cmd

	EnableKVM     bool
	CPU           string
	Memory        string
	NumProcessors int
	KernelFile    string
}

func (vm *QemuVM) Start() error {
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
		"-drive", "id=root,file=/home/colin/projects/layerfiles/prepare-disks/ubuntu-22.04.qcow2,format=qcow2,if=none",
		"-device", "virtio-blk-device,drive=root",
		"-append", "console=hvc0 root=/dev/vda rw acpi=off reboot=t panic=-1",
		//"-drive", "id=test,file=test.img,format=raw,if=none",
		//"-device", "virtio-blk-device,drive=test",
		//"-netdev", "tap,id=tap0,script=no,downscript=no",
		//"-device", "virtio-net-device,netdev=tap0",
	)

	//debug
	vm.Cmd.Stdout = os.Stdout
	vm.Cmd.Stderr = os.Stderr
	vm.Cmd.Stdin = os.Stdin

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

//func (vm *QemuVM) Save() error {
//
//}
//
//func (vm *QemuVM) Load(snap *snapshot.Snapshot) error {
//
//}
