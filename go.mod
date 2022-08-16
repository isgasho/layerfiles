module github.com/webappio/layerfiles

require (
	github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f
	github.com/pkg/errors v0.9.1
	github.com/schollz/progressbar/v3 v3.9.0
	github.com/webappio/layerfiles/pkg/vm_protocol v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab
)

replace github.com/webappio/layerfiles/pkg/vm_protocol => ./pkg/vm_protocol

go 1.16
