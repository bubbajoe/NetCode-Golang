package main

import (
	"fmt"
	"encoding/json"
)

type Node struct {
	Text string `json:"text"`
}

type File struct {
	Node
}

type Folder struct {
	Node
	Nodes []Node
}

// Folder 

func main() {
	a := &Folder{
		Folder{Text:"Folder 1"},
		Child: &File{Text:"File 1"},
	}
	str, _ := json.Marshal(a)
	fmt.Println(str)
}