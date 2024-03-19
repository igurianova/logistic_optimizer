package main

import (
	"github.com/igurianova/logistic_optimizer/internal/app"
	"log"
	"os"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
