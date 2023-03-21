package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	bc "github.com/antavelos/blockchain"

	"github.com/gin-gonic/gin"
)

const rootPort = "8080"
const rootHost = "http://localhost:8080"

var blockchain *bc.Blockchain

func initBlockchain() *bc.Blockchain {
	blockchain := bc.Blockchain{}

	blockchain.CreateGenesisBlock()

	// var users []string
	// for i := 0; i < 10; i++ {
	// 	users = append(users, bc.NewUuid())
	// }

	// for _, user := range users {
	// 	blockchain.AddTx(bc.Transaction{
	// 		Id:        NewUuid(),
	// 		Sender:    god,
	// 		Recipient: user,
	// 		Amount:    10.0,
	// 	})
	// }

	return &blockchain
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/blockchain", bc.ApiGetChain)
	router.POST("/transactions", bc.ApiAddTx)
	router.POST("/mine", bc.ApiMine)
	router.POST("/ping", bc.ApiPing)

	return router
}

func ping(selfPort string) {
	host := fmt.Sprintf("%v/ping", rootHost)
	selfHost := fmt.Sprintf("http://localhost:%v", selfPort)

	postBody, _ := json.Marshal(bc.Node{Host: selfHost})

	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(host, "application/json", responseBody)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var nodes []bc.Node
	json.Unmarshal(body, &nodes)

	for _, node := range nodes {
		if node.Host != selfHost {
			bc.DbAddNode(node)
		}
	}
	bc.DbAddNode(bc.Node{Host: rootHost})
}

func getChain(host string) *bc.Blockchain {
	url := host + "/blockchain"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var blockchain bc.Blockchain
	if err := json.Unmarshal(body, &blockchain); err != nil {
		log.Fatal("ooopsss! an error occurred, please try again")
	}

	if err != nil {
		log.Fatal("ooopsss an error occurred, please try again")
	}
	defer resp.Body.Close()

	return &blockchain
}

func main() {
	port := flag.String("port", "8080", "the nodes port")
	flag.Parse()

	if *port != rootPort {
		ping(*port)
	}

	if !bc.DbBlockchainExists() {
		if *port == rootPort {
			blockchain = initBlockchain()
			bc.DbSaveBlockchain(*blockchain)
		} else {
			var blockchain bc.Blockchain

			nodes, _ := bc.DbLoadNodes()
			for _, node := range nodes {
				nodeBlockchain := getChain(node.Host)
				if len(nodeBlockchain.Blocks) > len(blockchain.Blocks) {
					blockchain = *nodeBlockchain
				}
			}
			bc.DbSaveBlockchain(blockchain)
		}
	}

	router := initRouter()
	router.Run(fmt.Sprintf("localhost:%v", *port))
}
