package main

import (
	"log"
	"os"
)

func Good() {
	log.Println("ok")
}

func Bad1() {
	log.Fatal("bad") // want `log.Fatal/os.Exit not allowed outside main`
}

func Bad2() {
	os.Exit(1) // want `log.Fatal/os.Exit not allowed outside main`
}
