package load

import (
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"strings"
)

// Stock TypeScript declares captures as string even when they did not
// participate. Correct that type before the checker proves program types.
type regexpLibraryFS struct{ vfs.FS }

func (s *regexpLibraryFS) ReadFile(path string) (string, bool) {
	text, ok := s.FS.ReadFile(path)
	if strings.HasSuffix(path, "/lib.es5.d.ts") {
		// ECMAScript permits an omitted pattern; the stock declaration omits
		// these overloads although lowering already compiles the empty pattern.
		text = strings.ReplaceAll(text, "interface RegExpConstructor {", "interface RegExpConstructor {\n    new (): RegExp;\n    (): RegExp;\n    new (pattern: undefined, flags?: string): RegExp;\n    (pattern: undefined, flags?: string): RegExp;")
		text = strings.ReplaceAll(text, "interface RegExpExecArray extends Array<string>", "interface RegExpExecArray extends Array<string | undefined>")
		text = strings.ReplaceAll(text, "interface RegExpMatchArray extends Array<string>", "interface RegExpMatchArray extends Array<string | undefined>")
		text = strings.ReplaceAll(text, "split(separator: string | RegExp, limit?: number): string[];", "split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];")
	}
	if strings.HasSuffix(path, "/lib.es2018.regexp.d.ts") {
		text = strings.ReplaceAll(text, "[key: string]: string", "[key: string]: string | undefined")
	}
	if strings.HasSuffix(path, "/lib.es2015.symbol.wellknown.d.ts") {
		text = strings.ReplaceAll(text, "[Symbol.split](string: string, limit?: number): string[]", "[Symbol.split](string: string, limit?: number): (string | undefined)[]")
		text = strings.ReplaceAll(text, "split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];", "split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];")
	}
	return text, ok
}
