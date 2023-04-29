package node

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
}

func NewNode(name string, ip string, port string) Node {
	return Node{
		Name:   name,
		Schema: "http",
		IP:     ip,
		Port:   port,
	}
}

func (n *Node) Update(updated Node) {
	n.IP = updated.IP
	n.Port = updated.Port
}

func (n Node) GetHost() string {
	return fmt.Sprintf("%v://%v:%v", n.Schema, n.IP, n.Port)
}

func Unmarshal(data []byte) (node Node, err error) {
	err = json.Unmarshal(data, &node)
	return
}

func UnmarshalMany(data []byte) (nodes []Node, err error) {
	err = json.Unmarshal(data, &nodes)
	return
}

func AddNode(nodes []Node, newNode Node) ([]Node, error) {

	index := containsNode(nodes, newNode)
	if index == -1 {
		nodes = append(nodes, newNode)
	} else {
		nodes[index].Update(newNode)
	}

	return nodes, nil
}

func containsNode(nodes []Node, node Node) int {
	for i, n := range nodes {
		if n.Name == node.Name {
			return i
		}
	}
	return -1
}
