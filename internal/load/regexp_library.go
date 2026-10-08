package load

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Stock TypeScript declares captures as string even when they did not
// participate. Correct that type before the checker proves program types.
type regexpLibraryFS struct{ vfs.FS }

func (s *regexpLibraryFS) ReadFile(path tspath.RootedFilePath) (string, bool) {
	text, ok := s.FS.ReadFile(path)
	if strings.HasSuffix(path.AsString(), "/lib.es5.d.ts") || strings.HasSuffix(path.AsString(), "/lib.es2021.string.d.ts") || strings.HasSuffix(path.AsString(), "/lib.es2015.symbol.wellknown.d.ts") {
		// Replacement callbacks undergo ToString at runtime, independent of their return type.
		text = strings.ReplaceAll(text, "replacer: (substring: string, ...args: any[]) => string", "replacer: (substring: string, ...args: any[]) => unknown")
	}
	if strings.HasSuffix(path.AsString(), "/lib.es5.d.ts") {
		text = strings.ReplaceAll(text, "interface RegExpExecArray extends Array<string>", "interface RegExpExecArray extends Array<string | undefined>")
		text = strings.ReplaceAll(text, "interface RegExpMatchArray extends Array<string>", "interface RegExpMatchArray extends Array<string | undefined>")
		text = strings.ReplaceAll(text, "split(separator: string | RegExp, limit?: number): string[];", "split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];")
	}
	if strings.HasSuffix(path.AsString(), "/lib.es2018.regexp.d.ts") {
		text = strings.ReplaceAll(text, "[key: string]: string", "[key: string]: string | undefined")
	}
	if strings.HasSuffix(path.AsString(), "/lib.es2015.symbol.wellknown.d.ts") {
		text = strings.ReplaceAll(text, "[Symbol.split](string: string, limit?: number): string[]", "[Symbol.split](string: string, limit?: number): (string | undefined)[]")
		text = strings.ReplaceAll(text, "split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];", "split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];")
	}
	return text, ok
}
