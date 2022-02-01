package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestProcess(t *testing.T) {

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = Process(root, ".")
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Join(results, "\n"))

}
