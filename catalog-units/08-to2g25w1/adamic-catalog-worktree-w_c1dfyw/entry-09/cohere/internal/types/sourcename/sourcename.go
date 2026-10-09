// Package sourcename says what a source file's name means to cohere's TypeScript decisions.
//
// Adamic's `.a` files are TypeScript source under their own names (#6mhafvb): a program whose tsconfig
// lists ".a" in its top-level "sourceExtensions" holds X.a as X.a, never as an X.a.ts alias, so every
// finding, write and cache names the real file. Every decision cohere makes on a name's `.ts` ending
// (a rule's extension check, the printer's `<T,>` comma, the fix guard's parser) asks TreatedAs, so an
// `.a` file is decided exactly as the same file named `.ts` would be. The test beside this file holds
// every `.ts` decision in the module to it.
//
// The mapping is a name's, not a program's: TreatedAs says how to treat a file once it is TypeScript
// source. Whether a `.a` file is source at all is the program's to say, since `.a` is also the
// static-library extension; a phase that meets files the program has not claimed, the format walk,
// asks the program first.
package sourcename

import "strings"

// AdamicExtension is Adamic's own source extension.
const AdamicExtension = ".a"

// TreatedAs is the name a file's TypeScript decisions read: an Adamic `.a` name with `.ts` in place of
// its extension, so `Dog.test.a` reads as `Dog.test.ts`, and any other name as it is.
func TreatedAs(fileName string) string {
	if stem, isAdamic := strings.CutSuffix(fileName, AdamicExtension); isAdamic && stem != "" && !strings.HasSuffix(stem, "/") {
		return stem + ".ts"
	}
	return fileName
}

// IsAdamic reports whether a name is an Adamic `.a` source name.
func IsAdamic(fileName string) bool {
	return TreatedAs(fileName) != fileName
}
