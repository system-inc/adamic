// Read-only compatibility probe for the named shared registry's numeric kinds.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"os"
)

func main() {
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		var descriptor registry.Descriptor
		err = json.Unmarshal(data, &descriptor)
		fmt.Printf("%s: %v\n", path, err)
	}
}
