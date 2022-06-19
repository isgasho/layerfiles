all: lexer main

GO=/usr/local/go/bin/go
JAVA=/usr/bin/java
ANTLR=/usr/local/lib/antlr-4.7-complete.jar

lexer pkg/Layerfile.g4:
	$(JAVA) -jar $(ANTLR) -Dlanguage=Go -package lexer pkg/layerfile/Layerfile.g4
	rm pkg/layerfile/Layerfile.tokens pkg/layerfile/Layerfile.interp
	mkdir -p pkg/layerfile/lexer
	mv pkg/layerfile/layerfile_lexer.go pkg/layerfile/lexer/

main:
	$(GO) build -o lf pkg/main/main.go
