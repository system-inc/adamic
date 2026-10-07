package main

import (
	"encoding/json"
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strconv"
	"strings"
)

func main() {
	definitions := tailwind.AdamicWave06Definitions()
	if os.Args[1] == "definitions" {
		fmt.Print(string(definitions))
		return
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var input struct{ Cases []string }
	if err := json.Unmarshal(data, &input); err != nil {
		panic(err)
	}
	for _, name := range input.Cases {
		reading, found := tailwind.FrameworkStaticReading(name)
		order := make([]string, 0, len(reading.Order))
		for _, v := range reading.Order {
			order = append(order, strconv.Itoa(v))
		}
		shared, independent := tailwind.AdamicWave06Aliases(reading, found)
		fmt.Printf("%t %t %d %s %t %t\n", found, reading.Order == nil, reading.Count, strings.Join(order, ","), shared, independent)
	}
}
