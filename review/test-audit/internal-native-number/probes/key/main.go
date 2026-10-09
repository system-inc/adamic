package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
)

func main() {
	root, _ := os.Getwd()
	a, e := buildcache.Key(root, buildcache.Inputs{Name: "alpha"})
	if e != nil {
		panic(e)
	}
	b, e := buildcache.Key(root, buildcache.Inputs{Name: "beta"})
	if e != nil {
		panic(e)
	}
	fmt.Printf("alpha=%s\nbeta=%s\nequal=%t\n", a, b, a == b)
}
