//go:build ignore
// +build ignore

package main

import (
	"log"

	"github.com/adnaan/fir"
)

func main() {
	if err := fir.GeneratePublic(fir.InDir("../"), fir.OutDir("../public")); err != nil {
		log.Fatalf("failed generating public directory: %v", err)
	}
}
