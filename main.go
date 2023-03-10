package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/pboyd/antisopsephy/internal/isopsephy"
	"github.com/pboyd/antisopsephy/internal/lgpn"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <number>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Finds Greek names encoded as a number using Isopsephy.\n")
		os.Exit(1)
	}

	number, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid number: %s\n", err)
		os.Exit(1)
	}

	names, err := lgpn.Names(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to retrieve names: %v\n", err)
		os.Exit(1)
	}

	for name := range names {
		n, err := isopsephy.Calculate(name)
		if err != nil {
			continue
		}

		if n == number {
			fmt.Println(name)
		}
	}
}
