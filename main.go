package main

import (
	"log"
	"os"

	"github.com/toszyrt/opossum/cmd"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd.Execute()
}
