all: main

GO=/usr/local/go/bin/go
JAVA=/usr/bin/java
ANTLR=/usr/local/lib/antlr-4.7-complete.jar

lexer pkg/Layerfile.g4:
	$(JAVA) -jar $(ANTLR) -Dlanguage=Go -package lexer pkg/layerfile/Layerfile.g4
	rm pkg/layerfile/Layerfile.tokens pkg/layerfile/Layerfile.interp
	mkdir -p pkg/layerfile/lexer
	mv pkg/layerfile/layerfile_lexer.go pkg/layerfile/lexer/

filewatcher-proto pkg/fuse-filewatcher/vm_protocol_model/FuseMessage.proto pkg/fuse-filewatcher/vm_protocol_model/MetaMessage.proto:
	cd pkg/vm_protocol/vm_protocol_model && \
	protoc -I=. --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  *.proto;

pkg/qemu/qemu-system-x86_64:
	# qemu v7.0.0 with:
	# ../configure --disable-glusterfs --disable-seccomp --disable-{bzip2,snappy,lzo} --disable-usb-redir --disable-libusb --disable-libnfs --disable-libiscsi --disable-rbd  --disable-spice --disable-attr         --disable-cap-ng --disable-linux-aio --disable-brlapi         --disable-vnc-{jpeg,sasl,png} --disable-rdma --disable-curl --disable-curses --disable-sdl --disable-gtk  --disable-tpm --disable-vte --disable-vnc --disable-xen --disable-opengl --target-list=x86_64-softmmu
	cp ~/projects/qemu/build/qemu-system-x86_64 pkg/qemu/qemu-system-x86_64

pkg/qemu/qboot.rom:
	# see qemu-system-x86_64
	cp ~/projects/qemu/pc-bios/qboot.rom pkg/qemu/qboot.rom

pkg/vm/qemu-img:
	cp ~/projects/qemu/build/qemu-img pkg/vm/qemu-img

main: lexer filewatcher-proto pkg/qemu/qemu-system-x86_64 pkg/qemu/qboot.rom pkg/vm/qemu-img
	$(GO) build -o lf pkg/main/main.go
