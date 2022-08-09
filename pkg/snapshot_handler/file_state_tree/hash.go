package file_state_tree

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"sort"
	"syscall"
)

func ensureNodeOpened(node *Node) error {
	if node.FileDescriptor != 0 {
		return nil
	}
	if node.IsRoot() {
		var err error
		node.FileDescriptor, err = syscall.Open(node.GetPath(), 0600, syscall.O_RDONLY)
		return errors.Wrapf(err, "could not open(/%v)", node.Name)
	}

	err := ensureNodeOpened(node.Parent)
	if err != nil {
		return err
	}
	node.FileDescriptor, err = syscall.Openat(node.Parent.FileDescriptor, node.Name, syscall.O_RDONLY, 0600)
	return errors.Wrapf(err, "could not openat(%v, %v)", node.Parent.FileDescriptor, node.Name)
}

func createFileFromNode(node *Node) (*os.File, error) {
	err := ensureNodeOpened(node)
	if err != nil {
		return nil, err
	}
	//we have to dup because go closes fds on the NewFile finalizer
	newFd, err := syscall.Dup(node.FileDescriptor)
	if err != nil {
		return nil, errors.Wrapf(err, "could not duplicate dup(%v)", node.FileDescriptor)
	}
	f := os.NewFile(uintptr(newFd), node.Name)
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, errors.Wrap(err, "could not SEEK(0, SEEK_START)")
	}
	return f, nil
}

func readLinkAt(parentFD int, name string) (string, error) {
	for i := 128; ; i *= 2 {
		out := make([]byte, i)
		n, err := unix.Readlinkat(parentFD, name, out)
		if err != nil {
			return "", errors.Wrapf(err, "could not readlinkat(%v, %v)", parentFD, name)
		}
		if n < i {
			return string(out[:n]), nil
		}
	}
}

func hashSymlink(node *Node) (string, error) {
	var linkDest string
	var err error
	if node.IsRoot() {
		linkDest, err = os.Readlink(node.GetPath())
		if err != nil {
			return "", errors.Wrapf(err, "could not readlink(%v)", node.Name)
		}
	} else {
		err := ensureNodeOpened(node.Parent)
		if err != nil {
			return "", err
		}
		linkDest, err = readLinkAt(node.Parent.FileDescriptor, node.Name)
		if err != nil {
			return "", err
		}
	}
	hash := sha1.Sum([]byte(linkDest))
	return hex.EncodeToString(hash[:]), nil
}

func hashRegularFile(node *Node) (string, error) {
	f, err := createFileFromNode(node)
	if err != nil {
		return "", err
	}

	hasher := sha1.New()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return "", errors.Wrapf(err, "could not hash regular file at %v (%v)", uintptr(node.FileDescriptor), node.Name)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashDirectory(node *Node) (string, error) {
	f, err := createFileFromNode(node)
	if err != nil {
		return "", err
	}

	names, err := f.Readdirnames(-1)
	if err != nil {
		return "", errors.Wrapf(err, "could not list directory %v with file descriptor %v", node.Name, node.FileDescriptor)
	}

	sort.Strings(names)

	hasher := sha1.New()
	for _, name := range names {
		hasher.Write([]byte(name))
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (node *Node) SetHashFromContent() error {
	var fileInfo unix.Stat_t

	if node.IsRoot() {
		err := unix.Lstat(node.GetPath(), &fileInfo)
		if err != nil {
			return errors.Wrapf(err, "could not lstat(/%v)", node.Name)
		}
	} else {
		err := ensureNodeOpened(node.Parent)
		if err != nil {
			return err
		}
		err = unix.Fstatat(node.Parent.FileDescriptor, node.Name, &fileInfo, unix.AT_SYMLINK_NOFOLLOW)
		if err != nil {
			return errors.Wrapf(err, "could not fstatat(%v, %v)", node.Parent.FileDescriptor, node.Name)
		}
	}

	var err error
	switch fileInfo.Mode & unix.S_IFMT {
	case unix.S_IFDIR:
		node.ContentHash, err = hashDirectory(node)
		return err
	case unix.S_IFLNK:
		node.ContentHash, err = hashSymlink(node)
		return err
	case unix.S_IFREG:
		node.ContentHash, err = hashRegularFile(node)
		return err
	case unix.S_IFBLK, unix.S_IFCHR:
		return fmt.Errorf("cannot hash block object named %v", node.Name)
	default:
		return fmt.Errorf("cannot hash strange object %v of mode %v", node.Name, fileInfo.Mode)
	}
}
