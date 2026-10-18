// This probe observes loader behavior without lowering or native emission.
package main

import (
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"os"
)

func main() {
	program, err := load.Load(os.Args[1:])
	result := map[string]any{"loaded": err == nil}
	if err != nil {
		result["error"] = err.Error()
	} else {
		result["entries"] = len(program.Files())
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
