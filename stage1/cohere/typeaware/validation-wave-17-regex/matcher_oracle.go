package main

import (
	"fmt"
	"regexp"
)

func main() {
	pattern := regexp.MustCompile("^(is|has)[A-Z]([A-Za-z0-9]?)+")
	names := []string{"", "is", "has", "isA", "hasA", "isEnabled", "hasChildren", "enabled", "IsEnabled", "is_Enabled", "isÄ", "is🌍", "isA🌍", "hasA\n", " isEnabled", "isA-b", "isA0"}
	for _, name := range names {
		fmt.Println(pattern.MatchString(name))
	}
	for code := 0; code < 128; code++ {
		letter := string(rune(code))
		fmt.Println(pattern.MatchString("is" + letter))
		fmt.Println(pattern.MatchString("has" + letter + "x"))
		fmt.Println(pattern.MatchString("was" + letter))
	}
}
