package main

import (
	"encoding/json"
	"os"
	"unicode"
)

func main() {
	rows := [][]uint32{}
	for _, r := range unicode.Upper.R16 {
		rows = append(rows, []uint32{uint32(r.Lo), uint32(r.Hi), uint32(r.Stride)})
	}
	for _, r := range unicode.Upper.R32 {
		rows = append(rows, []uint32{r.Lo, r.Hi, r.Stride})
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"version": unicode.Version, "ranges": rows})
}
