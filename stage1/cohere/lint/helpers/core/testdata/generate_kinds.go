// Generate the parser-name to Go ast.Kind table from the checked-in gitlink.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/coherepin"
)

func main() {
	root := os.Args[1]
	pin, err := coherepin.Pinned(root)
	if err != nil {
		panic(err)
	}
	if err = coherepin.Check(root, pin); err != nil {
		panic(err)
	}
	fmt.Printf("// Generated from pinned Go ast.Kind (%s). Regenerate with testdata/generate_kinds.go.\n", pin[:8])
	fmt.Println("const kinds: readonly string[] = [")
	for kind := ast.Kind(0); kind <= ast.KindCount; kind++ {
		fmt.Printf("    '%s',\n", strings.TrimPrefix(kind.String(), "Kind"))
	}
	fmt.Println("];\nexport function kindNumber(kind: string): number { return kinds.indexOf(kind); }")
}
