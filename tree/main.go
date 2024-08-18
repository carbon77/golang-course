package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

type Node struct {
	Name  string
	IsDir bool
	Nodes []*Node
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	node, err := getNode(path, printFiles)
	if err != nil {
		return err
	}

	for idx, subNode := range node.Nodes {
		isLast := idx == len(node.Nodes)-1
		if err := subNode.Print(out, 0, 0, isLast); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) Print(out io.Writer, indent int, wallCount int, isLast bool) error {
	prefix := strings.Repeat("│\t", wallCount) + strings.Repeat("\t", indent-wallCount)
	if isLast {
		prefix += "└───"
	} else {
		prefix += "├───"
	}
	if _, err := out.Write([]byte(prefix + n.Name + "\n")); err != nil {
		return err
	}

	if n.IsDir {
		for idx, subNode := range n.Nodes {
			newWallCount := wallCount + 1
			if isLast {
				newWallCount -= 1
			}
			if err := subNode.Print(out, indent+1, newWallCount, idx == len(n.Nodes)-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func getNode(path string, printFiles bool) (node *Node, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	info, err := file.Stat()
	if err != nil {
		return
	}

	node = &Node{
		Name:  info.Name(),
		IsDir: info.IsDir(),
	}

	if !node.IsDir {
		sizeStr := fmt.Sprintf("(%db)", info.Size())
		if info.Size() == 0 {
			sizeStr = "(empty)"
		}
		node.Name += " " + sizeStr
		return
	}

	names, err := file.Readdirnames(-1)
	if err != nil {
		return
	}
	sort.Strings(names)

	for _, name := range names {
		subNode, err := getNode(path+string(os.PathSeparator)+name, printFiles)
		if err != nil {
			return nil, err
		}

		if subNode.IsDir || printFiles {
			node.Nodes = append(node.Nodes, subNode)
		}
	}

	return
}
