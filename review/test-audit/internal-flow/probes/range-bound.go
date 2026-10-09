package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/flow"
)

func main() {
	fmt.Printf("MutableRange{Start:1, End:3}.Contains(1) = %t\n", (flow.MutableRange{Start: 1, End: 3}).Contains(1))
}
