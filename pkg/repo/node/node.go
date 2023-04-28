package reponode

import (
	"encoding/json"

	database "github.com/antavelos/blockchain/pkg/db"
	nd "github.com/antavelos/blockchain/pkg/models/node"
)

type NodeRepo struct {
	db *database.DB
}

func NewNodeRepo(db *database.DB) *NodeRepo {
	return &NodeRepo{db: db}
}

func (r *NodeRepo) GetNodes() (nodes []nd.Node, err error) {
	data, err := r.db.Load()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &nodes)

	return nodes, err
}

func (r *NodeRepo) AddNode(node nd.Node) error {
	return r.db.WithLock(func(data []byte) (any, error) {
		nodes, _ := nd.UnmarshalMany(data)

		return nd.AddNode(nodes, node)
	})
}
