package main

import (
	"fmt"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	router := InitRouter()

	router.Run(fmt.Sprintf(":%v", port))
}
