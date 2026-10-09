// buildcache-publish runs as a separate gate unit after tests, while their local products still exist.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/buildcache"
)

func main() {
	audit := flag.Bool("audit", false, "drain sampled rebuild audits instead of uploads")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: buildcache-publish [-audit]")
		os.Exit(2)
	}
	drain := buildcache.PublishSpool
	if *audit {
		drain = buildcache.DrainAudits
	}
	if err := drain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
