// Independent Go expectations for graph-event decisions, not a production-rule
// byte oracle. The production byte oracle is oracle.go and covers the race rule.
package main

import "fmt"

func main() {
	for _, line := range []string{"2", "", "", "2", "", "2:4,5", "2:4", "2"} {
		fmt.Println(line)
	}
}
