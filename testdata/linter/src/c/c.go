package main

import (
	mylog "log"
	myos "os"
)

func BadWithAlias() {
	mylog.Fatal("bad") // want `log.Fatal/os.Exit not allowed outside main`
	myos.Exit(1)       // want `log.Fatal/os.Exit not allowed outside main`
}
