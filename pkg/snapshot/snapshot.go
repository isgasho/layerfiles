package snapshot

type FileNode struct {
	//ContentHash is the hash of the contents of this file or directory, empty if contents have not been read
	ContentHash string `json:"h,omitempty"`

	//Name is the name of this file or directory
	Name string
}

type Snapshot struct {
	UUID string `json:"uuid"`
	Root *FileNode
}
