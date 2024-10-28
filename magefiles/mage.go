//go:build mage

package main

import (
	"os"

	//mage:import
	golang "github.com/elisasre/mageutil/golang/target"
	//mage:import
	_ "github.com/elisasre/mageutil/git/target"
)

// Configure imported targets
func init() {
	os.Setenv("CGO_ENABLED", "0")
	golang.BuildTarget = "./cmd/golden-demo"
}
