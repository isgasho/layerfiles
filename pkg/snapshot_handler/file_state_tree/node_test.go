package file_state_tree

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func createFileWithContents(t *testing.T, args ...string) {
	if len(args) < 3 {
		t.Fatal("invalid usage: expected createFileWithContents(t, dir, ..., subdir, file, content)")
		return
	}

	f, err := os.OpenFile(filepath.Join(args[:len(args)-1]...), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
		return
	}

	_, err = f.WriteString(args[len(args)-1])
	if err != nil {
		t.Fatal(err)
		return
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
		return
	}
}

func createSubdir(t *testing.T, dir, subdir string) {
	err := os.Mkdir(filepath.Join(dir, subdir), 0700)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
}

func TestNodeFromPath(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "test-node-from-path-*")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer os.RemoveAll(tmpDir)

	createFileWithContents(t, tmpDir, "file1", "file 1 contents")
	createSubdir(t, tmpDir, "subdir")
	createFileWithContents(t, tmpDir, "subdir", "file2", "file 2 contents")

	node, err := TreeFromDir(tmpDir)
	if err != nil {
		t.Fatal(err)
		return
	}

	json, err := node.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if string(json) != `{"h":"d8f8b13a25c9109a288abad5ee1a3266c821b283","c":{"file1":{"h":"f9116ff67427bdaed292888c1a7d8bc664d115c0"},"subdir":{"h":"cb99b709a1978bd205ab9dfd4c5aaa1fc91c7523","c":{"file2":{"h":"16638beb18762c56af7674a16043aa64b92f693f"}}}}}` {
		t.Fatal("JSON was not what was expected, got " + string(json))
	}
}

func TestAddNodes(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", "test-node-from-path-*")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer os.RemoveAll(tmpDir)

	createFileWithContents(t, tmpDir, "file1", "file 1 contents")
	createSubdir(t, tmpDir, "subdir")
	createFileWithContents(t, tmpDir, "subdir", "file2", "file 2 contents")

	node := &Node{Name: tmpDir}

	err = node.NodeFromPath("file1").SetHashFromContent()
	if err != nil {
		t.Fatal(err)
		return
	}

	json, err := node.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	if string(json) != `{"c":{"file1":{"h":"f9116ff67427bdaed292888c1a7d8bc664d115c0"}}}` {
		t.Fatal("JSON was not what was expected, got " + string(json))
	}

	err = node.NodeFromPath("subdir/file2").SetHashFromContent()
	if err != nil {
		t.Fatal(err)
		return
	}

	json, err = node.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	if string(json) != `{"c":{"file1":{"h":"f9116ff67427bdaed292888c1a7d8bc664d115c0"},"subdir":{"c":{"file2":{"h":"16638beb18762c56af7674a16043aa64b92f693f"}}}}}` {
		t.Fatal("JSON was not what was expected, got " + string(json))
	}

	err = node.SetHashFromContent()
	if err != nil {
		t.Fatal(err)
		return
	}

	err = node.NodeFromPath("subdir").SetHashFromContent()
	if err != nil {
		t.Fatal(err)
		return
	}

	json, err = node.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	//this is the same as the TreeFromJson test
	if string(json) != `{"h":"d8f8b13a25c9109a288abad5ee1a3266c821b283","c":{"file1":{"h":"f9116ff67427bdaed292888c1a7d8bc664d115c0"},"subdir":{"h":"cb99b709a1978bd205ab9dfd4c5aaa1fc91c7523","c":{"file2":{"h":"16638beb18762c56af7674a16043aa64b92f693f"}}}}}` {
		t.Fatal("JSON was not what was expected, got " + string(json))
	}
}
