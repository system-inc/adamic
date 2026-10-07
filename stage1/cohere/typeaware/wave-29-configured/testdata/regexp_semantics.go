// Cohere id-match uses Go regexp, not ECMAScript RegExp.
package main

import (
	"fmt"
	"regexp"
)

func main() {
	for index, input := range [][2]string{{`\Afoo\z`, "foo"}, {`^foo$`, "foo\n"}, {`.`, "\r"}, {`(?i)^foo$`, "FOO"}, {`^[^_]+$`, "foo"}} {
		pattern, err := regexp.Compile(input[0])
		if err != nil {
			fmt.Printf("%d invalid\n", index)
			continue
		}
		fmt.Printf("%d match %t\n", index, pattern.MatchString(input[1]))
	}
}
