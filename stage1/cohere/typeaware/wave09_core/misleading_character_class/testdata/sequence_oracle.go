package main

import (
	"encoding/json"
	"fmt"
	core "github.com/system-inc/cohere/internal/lint/rules/core"
	"os"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var fixture struct {
		Rows     [][]core.Wave09ClassCharacter
		Patterns []string
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		panic(err)
	}
	for at, row := range fixture.Rows {
		fmt.Printf("case %d\n", at)
		fmt.Print(core.Wave09Sequence(row))
	}
	for _, pattern := range fixture.Patterns {
		fmt.Println(core.Wave09MeaningChanges(pattern))
	}
}
