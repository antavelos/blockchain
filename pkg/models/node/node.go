package node

import (
	"encoding/json"
	"fmt"

	"github.com/antavelos/blockchain/pkg/lib/rest"
)

type Node struct {
	Name   string `json:"name"`
	Schema string `json:"schema" default:"http"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
}

func (n Node) GetHost() string {
	return fmt.Sprintf("%v://%v:%v", n.Schema, n.IP, n.Port)
}

type NodeMarshaller rest.ObjectMarshaller

// TODO: make this generic
func (nm NodeMarshaller) Unmarshal(data []byte) (any, error) {
	var target any
	if nm.Many {
		target = make([]Node, 0)
	} else {
		target = Node{}
	}

	err := json.Unmarshal(data, &target)

	return target, err
}
