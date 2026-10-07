package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows []struct {
		tailwind.ClassLiteralSettings
		Key *string
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for _, row := range rows {
		key := tailwind.AdamicSettingsKey(row.ClassLiteralSettings)
		if row.Key != nil {
			key = *row.Key
		}
		identity, creates, settingsKey := tailwind.AdamicCompiledReader(key, row.ClassLiteralSettings)
		var bytes strings.Builder
		for _, b := range []byte(settingsKey) {
			fmt.Fprintf(&bytes, "%d,", b)
		}
		fmt.Printf("reader %d %d %s\n", identity, creates, bytes.String())
	}
}
