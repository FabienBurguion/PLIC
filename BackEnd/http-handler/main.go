package main

import (
	"PLIC/utils"
	"fmt"
)

func main() {
	utils.GET("/hello_world", GetHelloWorld)

	fmt.Println("Server running on port 8080...")
	utils.Start(":8080")
}
