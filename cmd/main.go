package main

import (
	"fmt"
	"os"
)

func loadDictTxt(path string) error {
	dict, _ := os.Open(path)
	fmt.Println(dict)
	return nil
}

func main() {
	loadDictTxt("../cfg/dict.txt")
}
