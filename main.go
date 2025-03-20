package main

import (
	"fmt"

	"github.com/zedd123/mlsolid/solid"
)

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	fmt.Println("mlsolid", config)
}
