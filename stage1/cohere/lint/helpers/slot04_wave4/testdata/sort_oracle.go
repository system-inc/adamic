package main

import (
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

type Entry struct {
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

func print(keys []string, label string) {
	var out strings.Builder
	out.WriteString(label)
	for _, key := range keys {
		out.WriteString(" [")
		for _, b := range []byte(key) {
			fmt.Fprintf(&out, "%d,", b)
		}
		out.WriteString("]")
	}
	fmt.Println(out.String())
}
func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows [][]Entry
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, row := range rows {
		set := map[string]bool{}
		for _, entry := range row {
			set[entry.Key] = entry.Value
		}
		keys := collapse.AdamicSortedKeys(set)
		print(keys, "keys")
		keys = append(keys, "mutate returned list")
		print(collapse.AdamicSortedKeys(set), "fresh")
		set["mutation after first call"] = false
		print(collapse.AdamicSortedKeys(set), "updated")
	}
}
