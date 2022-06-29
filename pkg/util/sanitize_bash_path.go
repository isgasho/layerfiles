package util

import "strings"

//SanitizeBashPath takes an arbitrary file name and sanitizes it such
//that you could "cat (file)" directly into a shell
//i.e., a file named "hello' world" would be catted as:
//cat 'hello'"'"' world'
func SanitizeBashPath(path string) string {
	path = strings.ReplaceAll(path, "'", `'\''`)
	return "'" + path + "'"
}
