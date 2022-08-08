all: main

GO=/usr/local/go/bin/go
JAVA=/usr/bin/java
ANTLR=/usr/local/lib/antlr-4.7-complete.jar

lexer pkg/Layerfile.g4:
	$(JAVA) -jar $(ANTLR) -Dlanguage=Go -package lexer pkg/layerfile/Layerfile.g4
	rm pkg/layerfile/Layerfile.tokens pkg/layerfile/Layerfile.interp
	mkdir -p pkg/layerfile/lexer
	mv pkg/layerfile/layerfile_lexer.go pkg/layerfile/lexer/

filewatcher-proto pkg/fuse-filewatcher/filewatcher_model/FuseMessage.proto pkg/fuse-filewatcher/filewatcher_model/MetaMessage.proto:
	cd pkg/fuse-filewatcher/filewatcher_model && \
	protoc -I=. --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  *.proto;

main: lexer filewatcher-proto
	$(GO) build -o lf pkg/main/main.go
