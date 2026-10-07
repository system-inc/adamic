package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var rows [][2]string
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	for i, row := range rows {
		pattern, err := regexp.Compile(row[0])
		if err != nil {
			panic(err)
		}
		result := 0
		if pattern.MatchString(row[1]) {
			result = 1
		}
		fmt.Printf("%d %d\n", i, result)
	}
}
