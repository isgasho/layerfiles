all: ubuntu-22.04.qcow2

clean:
	sudo rm -rf ubuntu-22.04 fuse-filewatcher-v2-linux-amd64

jammy-server-cloudimg-amd64-root-20220622.tar.xz:
	wget -O jammy-server-cloudimg-amd64-root-20220622.tar.xz https://cloud-images.ubuntu.com/jammy/20220622/jammy-server-cloudimg-amd64-root.tar.xz

fuse-filewatcher-v2-linux-amd64:
	cd ../pkg/fuse-filewatcher/client/main && \
	  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	  go build -mod=vendor -tags netgo -ldflags '-w' \
	  -o ../../../../prepare-disks/fuse-filewatcher-v2-linux-amd64

ubuntu-22.04: jammy-server-cloudimg-amd64-root-20220622.tar.xz fuse-filewatcher-v2-linux-amd64
	sudo rm -rf ubuntu-22.04
	mkdir ubuntu-22.04
	sudo tar -C ubuntu-22.04 -xJf jammy-server-cloudimg-amd64-root-20220622.tar.xz
	bash prepare-22.04-disk.sh
	sudo cp fuse-filewatcher-v2-linux-amd64 ubuntu-22.04/usr/local/bin/

ubuntu-22.04.qcow2: ubuntu-22.04
	sudo virt-make-fs --format=qcow2 --type=ext4 --size=60G ubuntu-22.04 ubuntu-22.04-large.qcow2
	qemu-img convert ubuntu-22.04-large.qcow2 -O qcow2 ubuntu-22.04.qcow2
	sudo rm ubuntu-22.04-large.qcow2
