package utils

import (
	"log"
	"os"
	"fmt"
)

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CheckIfRoot() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run with sudo or as root.")
		os.Exit(1)
	}
}
