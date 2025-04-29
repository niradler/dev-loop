package main

// @name: Hello (go)
// @description: A simple script that prints "Hello, {name}" to the console
// @author: Nir Adler
// @category: Testing
// @tags: ["hello", "test"]
// @inputs: [
//   { "name": "name", "description": "Your name", "type": "string", "default": "" }
// ]

import (
	"fmt"
	"os"
)

func main() {
	name := os.Args[1]
	fmt.Println("Hello,", name)
}
