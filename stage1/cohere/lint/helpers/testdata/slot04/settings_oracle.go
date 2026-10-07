package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []tailwind.ClassLiteralSettings
	if err := json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, row := range rows {
		for _, b := range []byte(tailwind.AdamicSettingsKey(row)) {
			fmt.Printf("%d,", b)
		}
		fmt.Println()
	}
}
