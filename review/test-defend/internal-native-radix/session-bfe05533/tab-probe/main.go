package main

import (
	"fmt"
	regex "github.com/system-inc/adamic/internal/regexp"
)

func main() {
	p, err := regex.Parse(`\t`, "")
	if err != nil {
		panic(err)
	}
	c := p.Body.Alternatives[0].Terms[0].(*regex.Character)
	fmt.Printf("parsed tab Value=%d Raw=%q\n", c.Value, c.Raw)
}
