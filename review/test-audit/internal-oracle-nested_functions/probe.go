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
	p, err := load.Load([]string{os.Args[1]})
	if err != nil {
		fmt.Printf("load-error: %v\n", err)
		return
	}
	q, err := lower.Lower(context.Background(), p)
	if err != nil {
		fmt.Printf("lower-error: %v\n", err)
		return
	}
	b, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
