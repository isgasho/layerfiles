package file_state_tree

import (
	"encoding/json"
	"strings"
	"syscall"
)

type Node struct {
	//ContentHash is the hash of the contents of this file or directory,
	//or "" if the file has not been read
	ContentHash string

	//Name is the name of this file
	Name string

	//FileDescriptor is the FD of this file, or 0 if the file has not
	//been read yet.
	FileDescriptor int

	//Children are the files contained in this directory
	Children map[string]*Node

	Parent *Node
}

func (node *Node) IsRoot() bool {
	return node.Parent == nil
}

func (node *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ContentHash string           `json:"h,omitempty"`
		Children    map[string]*Node `json:"c,omitempty"`
	}{
		ContentHash: node.ContentHash,
		Children:    node.Children,
	})
}

func (node *Node) UnmarshalJSON(data []byte) error {
	var unmarshalled struct {
		ContentHash string           `json:"h,omitempty"`
		Children    map[string]*Node `json:"c,omitempty"`
	}
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	node.Children = unmarshalled.Children
	node.ContentHash = unmarshalled.ContentHash
	if node.Children != nil {
		for name, child := range node.Children {
			child.Name = name
			child.Parent = node
		}
	}
	return nil
}

func (node *Node) AddChild(child *Node) {
	if node.Children == nil {
		node.Children = make(map[string]*Node)
	}
	node.Children[child.Name] = child
	child.Parent = node
}

func (node *Node) GetOrAddChild(name string) *Node {
	if node.Children == nil {
		node.Children = make(map[string]*Node)
	}

	if currChild, ok := node.Children[name]; ok {
		return currChild
	}

	node.Children[name] = &Node{Name: name, Parent: node}
	return node.Children[name]
}

func (node *Node) GetPath() string {
	if node.IsRoot() {
		return node.Name
	}
	return node.Parent.GetPath() + "/" + node.Name
}

func (node *Node) NodeFromPath(path string) *Node {
	if path == "" {
		return node
	}
	split := strings.SplitN(path, "/", 2)

	child := node.GetOrAddChild(split[0])

	if len(split) == 1 {
		return child
	}

	return child.NodeFromPath(split[1])
}

func (node *Node) Close() {
	if node.FileDescriptor != 0 {
		syscall.Close(node.FileDescriptor)
		node.FileDescriptor = 0
	}
	if node.Children != nil {
		for _, child := range node.Children {
			child.Close()
		}
	}
}
