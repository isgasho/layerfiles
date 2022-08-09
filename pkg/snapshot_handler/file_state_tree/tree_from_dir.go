package file_state_tree

import (
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"strings"
)

func fillNode(node *Node) error {
	var fileInfo unix.Stat_t
	var err error

	if node.IsRoot() {
		err = unix.Lstat(node.GetPath(), &fileInfo)
		if err != nil {
			return errors.Wrapf(err, "could not Lstat %v", node.Name)
		}
	} else {
		err = ensureNodeOpened(node.Parent)
		if err != nil {
			return err
		}
		err = unix.Fstatat(node.Parent.FileDescriptor, node.Name, &fileInfo, unix.AT_SYMLINK_NOFOLLOW)
		if err != nil {
			return errors.Wrapf(err, "could not Fstatat(%v, %v)", node.Parent.FileDescriptor, node.Name)
		}
	}

	err = node.SetHashFromContent()
	if err != nil {
		return err
	}

	if (fileInfo.Mode & unix.S_IFDIR) != 0 {
		f, err := createFileFromNode(node)
		if err != nil {
			return err
		}
		childrenNames, err := f.Readdirnames(-1)
		if err != nil {
			return errors.Wrapf(err, "could not list directory with file descriptor %v", node.FileDescriptor)
		}
		for _, childName := range childrenNames {
			if childName == ".git" {
				continue
			}
			child := &Node{
				Name:   childName,
				Parent: node,
			}
			err := fillNode(child)
			if err != nil {
				return err
			}
			node.AddChild(child)
		}
	}
	return nil
}

func TreeFromDir(dir string) (*Node, error) {
	dir = strings.TrimRight(dir, "/")
	root := &Node{ //i.e., "/"
		Name: dir,
	}
	return root, fillNode(root)
}
