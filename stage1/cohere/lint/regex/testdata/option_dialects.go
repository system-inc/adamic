package main

import (
	"fmt"
	"regexp"
)

func main() {
	for _, pair := range [][2]string{{`\s`, "\u00a0"}, {`a$`, "a\n"}, {`(?i)todo`, "TODO"}, {`\p{Greek}`, "α"}, {`a\z`, "a"}, {`(?P<word>a)`, "a"}} {
		fmt.Printf("%s\t%t\n", pair[0], regexp.MustCompile(pair[0]).MatchString(pair[1]))
	}
}
