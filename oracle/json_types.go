// json_types supplies the oracle's source loader with type descriptors, not lowered code.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
)

func main() {
	program, err := load.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sources, err := lower.DecodeJsonSources(context.Background(), program)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = json.NewEncoder(os.Stdout).Encode(sources); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
