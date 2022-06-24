all: ubuntu-22.04.qcow2

jammy-server-cloudimg-amd64-root-20220622.tar.xz:
	wget -O jammy-server-cloudimg-amd64-root-20220622.tar.xz https://cloud-images.ubuntu.com/jammy/20220622/jammy-server-cloudimg-amd64-root.tar.xz

ubuntu-22.04: jammy-server-cloudimg-amd64-root-20220622.tar.xz
	sudo rm -rf ubuntu-22.04
	mkdir ubuntu-22.04
	sudo tar -C ubuntu-22.04 -xJf jammy-server-cloudimg-amd64-root-20220622.tar.xz
	bash prepare-22.04-disk.sh

ubuntu-22.04.qcow2 ubuntu-22.04:
	sudo virt-make-fs --format=qcow2 --type=ext4 --size=60G ubuntu-22.04 ubuntu-22.04-large.qcow2
	qemu-img convert ubuntu-22.04-large.qcow2 -O qcow2 ubuntu-22.04.qcow2
	sudo rm ubuntu-22.04-large.qcow2