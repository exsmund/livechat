package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	defer os.Exit(0)

	file, err := os.OpenFile("logs.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	} else {
		defer file.Close()
		log.SetOutput(file)
	}

	a, err := NewApp()

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	a.Loop()
}
