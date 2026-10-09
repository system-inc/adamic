// buildcache-publish runs as a separate gate unit after tests, while their local products still exist.
package main

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/buildcache"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: buildcache-publish (audit draining is not implemented yet)")
		os.Exit(2)
	}
	if err := buildcache.PublishSpool(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
