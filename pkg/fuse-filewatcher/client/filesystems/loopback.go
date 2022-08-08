package filesystems

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"golang.org/x/sys/unix"
	"k8s.io/klog"
	"path/filepath"
	"sync"
	"syscall"
)

type LoopbackFileSystem struct {
	RootDir  string
	RootFD   int
	OnAccess func(path string)
	OnRead   func(path string)

	readsAllowed      bool
	readsAllowedMutex sync.Mutex
	readsAllowedCond  *sync.Cond
}

func (f *LoopbackFileSystem) Root() fs.InodeEmbedder {
	return &LoopbackNode{
		rootFS: f,
	}
}

func (f *LoopbackFileSystem) idFromStat(st *unix.Stat_t) fs.StableAttr {
	// We compose an inode number by the underlying inode, and
	// mixing in the device number. In traditional filesystems,
	// the inode numbers are small. The device numbers are also
	// small (typically 16 bit). Finally, we mask out the root
	// device number of the root, so a loopback FS that does not
	// encompass multiple mounts will reflect the inode numbers of
	// the underlying filesystem
	swapped := (uint64(st.Dev) << 32) | (uint64(st.Dev) >> 32)
	return fs.StableAttr{
		Mode: uint32(st.Mode),
		Gen:  1,
		// This should work well for traditional backing FSes,
		// not so much for other go-fuse FS-es
		Ino: swapped ^ st.Ino,
	}
}

func ToStatus(err error) fuse.Status {
	if err == nil {
		if klog.V(5) {
			klog.InfoDepth(1, "Success")
		}
	} else if klog.V(1) {
		klog.WarningDepth(1, err)
	}
	return fuse.ToStatus(err)
}

// Based on the hanwen/go-fuse loopback example, but wth "openat" everywhere
// (because we want to mount /a over /a in prod) - otherwise if you went by path
// instead of by fd, accessing /a would then request /a and cause an infinite loop.
func NewLoopbackFileSystem(root string) (*LoopbackFileSystem, error) {
	res := &LoopbackFileSystem{
		RootDir:  root,
		OnAccess: func(path string) {},
		OnRead:   func(path string) {},
	}
	res.readsAllowedCond = sync.NewCond(&res.readsAllowedMutex)
	return res, nil
}

type LoopbackNode struct {
	fs.Inode

	rootFS *LoopbackFileSystem
}

var _ = (fs.NodeStatfser)((*LoopbackNode)(nil))
var _ = (fs.NodeGetattrer)((*LoopbackNode)(nil))
var _ = (fs.NodeGetxattrer)((*LoopbackNode)(nil))
var _ = (fs.NodeListxattrer)((*LoopbackNode)(nil))
var _ = (fs.NodeReadlinker)((*LoopbackNode)(nil))
var _ = (fs.NodeOpener)((*LoopbackNode)(nil))
var _ = (fs.NodeLookuper)((*LoopbackNode)(nil))
var _ = (fs.NodeOpendirer)((*LoopbackNode)(nil))
var _ = (fs.NodeReaddirer)((*LoopbackNode)(nil))

func (n *LoopbackNode) waitReadsAllowed() {
	if !n.rootFS.readsAllowed {
		n.rootFS.readsAllowedMutex.Lock()
		for !n.rootFS.readsAllowed {
			n.rootFS.readsAllowedCond.Wait()
		}
		n.rootFS.readsAllowedMutex.Unlock()
	}
}

func (n *LoopbackNode) relpath() string {
	return n.Path(n.Root())
}

func (n *LoopbackNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	n.waitReadsAllowed()
	n.rootFS.OnAccess(n.relpath())

	s := syscall.Statfs_t{}

	err := syscall.Fstatfs(n.rootFS.RootFD, &s)
	if err == nil {
		out.FromStatfsT(&s)
		klog.V(5).Infof("StatFS(%v): %v", n.relpath(), "ok")
		return fs.OK
	}
	klog.Error("error with statFS(", n.relpath(), "): ", err)
	return fs.ToErrno(err)
}

func CopyAttrsFromUnixToFuse(st *unix.Stat_t, out *fuse.Attr) {
	out.Ino = st.Ino
	out.Size = uint64(st.Size)
	out.Blocks = uint64(st.Blocks)
	out.Atime = uint64(st.Atim.Sec)
	out.Atimensec = uint32(st.Atim.Nsec)
	out.Mtime = uint64(st.Mtim.Sec)
	out.Mtimensec = uint32(st.Mtim.Nsec)
	out.Ctime = uint64(st.Ctim.Sec)
	out.Ctime = uint64(st.Ctim.Nsec)
	out.Mode = st.Mode
	out.Nlink = uint32(st.Nlink)
	out.Uid = st.Uid
	out.Gid = st.Gid
	out.Rdev = uint32(st.Rdev)
}

func (n *LoopbackNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	name := n.relpath()

	st := unix.Stat_t{}

	if name == "" {
		// When GetAttr is called for the toplevel directory, we always want
		// to look through symlinks.
		err := unix.Fstat(n.rootFS.RootFD, &st)
		if err != nil {
			klog.V(5).Infof("GetAttr(%v): %v", name, err)
			return fs.ToErrno(err)
		}
	} else {
		n.waitReadsAllowed()

		n.rootFS.OnAccess(name)
		err := unix.Fstatat(n.rootFS.RootFD, name, &st, unix.AT_SYMLINK_NOFOLLOW)
		if err != nil {
			klog.V(5).Infof("GetAttr(%v): %v", name, err)
			return fs.ToErrno(err)
		}
	}

	CopyAttrsFromUnixToFuse(&st, &out.Attr)

	klog.V(5).Infof("GetAttr(%v): %v", name, "ok")
	return fs.OK
}

func (n *LoopbackNode) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnAccess(path)
	if path == "" {
		path = "."
	}

	fd, err := unix.Openat(n.rootFS.RootFD, path, 0, unix.O_PATH)
	if err != nil {
		klog.V(5).Infof("Getxattr(%v): %v", path, err)
		return 0, fs.ToErrno(err)
	}

	defer unix.Close(fd)

	sz, err := unix.Fgetxattr(fd, attr, dest)
	klog.V(5).Infof("Getxattr(%v): %v", path, err)
	return uint32(sz), fs.ToErrno(err)
}

func (n *LoopbackNode) Listxattr(ctx context.Context, dest []byte) (uint32, syscall.Errno) {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnAccess(path)
	if path == "" {
		path = "."
	}

	fd, err := unix.Openat(n.rootFS.RootFD, path, 0, unix.O_PATH)
	if err != nil {
		klog.V(5).Infof("Listxattr(%v): %v", path, err)
		return 0, fs.ToErrno(err)
	}
	defer unix.Close(fd)

	sz, err := unix.Flistxattr(fd, dest)
	klog.V(5).Infof("Listxattr(%v): %v", path, err)
	return uint32(sz), fs.ToErrno(err)
}

func (n *LoopbackNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnRead(path)
	if path == "" {
		path = "."
	}

	for l := 256; ; l *= 2 {
		buf := make([]byte, l)
		sz, err := unix.Readlinkat(n.rootFS.RootFD, path, buf)
		if err != nil {
			klog.V(5).Infof("Readlink(%v): %v", path, err)
			return nil, fs.ToErrno(err)
		}

		if sz < len(buf) {
			klog.V(5).Infof("Readlink(%v): ok", path)
			return buf[:sz], 0
		}
	}
}

func (n *LoopbackNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnRead(path)
	if path == "" {
		path = "."
	}

	flags = flags &^ syscall.O_APPEND
	f, err := syscall.Openat(n.rootFS.RootFD, path, int(flags), 0) //mode is 0: node must already exist
	if err != nil {
		klog.V(5).Infof("Open(%v): %v", path, err)
		return nil, 0, fs.ToErrno(err)
	}
	lf := fs.NewLoopbackFile(f)
	klog.V(5).Infof("Open(%v): ok", path)
	return lf, 0, 0
}

func (n *LoopbackNode) Opendir(ctx context.Context) syscall.Errno {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnRead(path)
	if path == "" {
		path = "."
	}

	fd, err := syscall.Openat(n.rootFS.RootFD, path, syscall.O_DIRECTORY, 0755)
	if err != nil {
		klog.V(5).Infof("Opendir(%v): %v", path, err)
		return fs.ToErrno(err)
	}
	syscall.Close(fd)
	klog.V(5).Infof("Opendir(%v): ok", path)
	return fs.OK
}

func (n *LoopbackNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	n.waitReadsAllowed()
	path := n.relpath()
	n.rootFS.OnRead(path)
	if path == "" {
		path = "."
	}

	klog.V(5).Infof("Readdir(%v): ok, rootdir=%v", path, n.rootFS.RootDir)
	return fs.NewLoopbackDirStream(filepath.Join(n.rootFS.RootDir, path)) //TODO doesn't allow src=dest
}

func (n *LoopbackNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	dirPath := n.relpath()
	p := name
	if dirPath != "" {
		p = filepath.Join(dirPath, name)
	}
	n.waitReadsAllowed()
	n.rootFS.OnRead(p)

	st := unix.Stat_t{}
	err := unix.Fstatat(n.rootFS.RootFD, p, &st, unix.AT_SYMLINK_NOFOLLOW)
	if err != nil {
		klog.V(5).Infof("Lookup(%v): %v", p, err)
		return nil, fs.ToErrno(err)
	}

	node := &LoopbackNode{
		rootFS: n.rootFS,
	}
	ch := n.NewInode(ctx, node, n.rootFS.idFromStat(&st))
	klog.V(5).Infof("Lookup(%v): %#v", p, ch)
	CopyAttrsFromUnixToFuse(&st, &out.Attr)

	return ch, 0
}

func (fs *LoopbackFileSystem) AllowReads() {
	fs.readsAllowedMutex.Lock()
	if !fs.readsAllowed {
		var err error
		fs.RootFD, err = syscall.Open(fs.RootDir, syscall.O_RDONLY, 0)
		if err != nil {
			klog.Error(err)
		}
		klog.V(3).Infof("Got an AllowReads request, allowing reads.")
		fs.readsAllowed = true
		fs.readsAllowedCond.Broadcast()
	}
	fs.readsAllowedMutex.Unlock()
}
