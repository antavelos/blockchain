package main

import "fmt"

func main() {
	router := initRouter()
	router.Run(fmt.Sprintf(":%v", 5000))
}
