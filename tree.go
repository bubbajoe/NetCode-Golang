package main

import (
	"fmt"
)

type Tree struct {
	Folder []Folder
	File []File
}

// {text:"Name",nodes:[]}
type Folder struct {
	Name string
	Node Tree
}

// {text:"Name"}
type File struct {
	Name string
	Data []byte
	Options []string
}

func NewTree() (*Tree) {
	return new(Tree)
}

func (t Tree) ToJSON() string {
	json := ""
	node := t
	for {
		for _, folder := range node.Folder {
			fp, fi := parse()
		}
		for _, file := range node.File {
			
			
		}
		break
	}
	return json
}

func parse(node Tree) ([]Folder, []File]) {
	return (node.Folder, node.file)
}

func main() {
	a := NewTree()
	b := a.ToJSON()
	fmt.Println(b)
}