module github.com/webappio/layerfiles

require (
	github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f
	github.com/pkg/errors v0.9.1
	github.com/webappio/layerfiles/pkg/fuse-filewatcher v0.0.0-00010101000000-000000000000
)

replace github.com/webappio/layerfiles/pkg/fuse-filewatcher => ./pkg/fuse-filewatcher

go 1.16
