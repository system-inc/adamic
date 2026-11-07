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
	program, loadError := load.Load(os.Args[1:])
	if loadError != nil {
		fmt.Fprintln(os.Stderr, loadError)
		os.Exit(1)
	}
	sources, decodeJsonSourcesError := lower.DecodeJsonSources(context.Background(), program)
	if decodeJsonSourcesError != nil {
		fmt.Fprintln(os.Stderr, decodeJsonSourcesError)
		os.Exit(1)
	}
	if encodeError := json.NewEncoder(os.Stdout).Encode(sources); encodeError != nil {
		fmt.Fprintln(os.Stderr, encodeError)
		os.Exit(1)
	}
}
