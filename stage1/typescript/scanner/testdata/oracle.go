// Built through an overlay inside typescript-go, so its internal scanner is the oracle.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

func written(text string) string {
	var out strings.Builder
	// Scanner values can contain CESU-8 lone surrogates. Preserve their code units.
	for i := 0; i < len(text); {
		var r rune
		var size int
		if i+2 < len(text) && text[i] == 0xed && text[i+1] >= 0xa0 && text[i+1] <= 0xbf && text[i+2]&0xc0 == 0x80 {
			r = rune(text[i]&15)<<12 | rune(text[i+1]&63)<<6 | rune(text[i+2]&63)
			size = 3
		} else {
			r, size = utf8.DecodeRuneInString(text[i:])
		}
		i += size
		if r >= 32 && r <= 126 && r != '\\' {
			out.WriteRune(r)
		} else if r <= 0xffff {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			high, low := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, high, low)
		}
	}
	return out.String()
}
func run(out *bufio.Writer, path, mode string, countOnly bool) int {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	s := scanner.NewScanner()
	s.SetText(string(text))
	if !countOnly {
		s.SetOnError(func(message *diagnostics.Message, start, length int, args ...any) {
			fmt.Fprintf(out, "error %d %d %d\n", message.Code(), start, length)
		})
	}
	count := 0
	for {
		var kind ast.Kind
		if mode == "jsx" {
			kind = s.ScanJsxToken()
		} else {
			kind = s.Scan()
		}
		if mode == "regex" {
			kind = s.ReScanSlashToken(false)
		}
		if mode == "greater" {
			kind = s.ReScanGreaterThanToken()
		}
		if mode == "template" && kind == ast.KindCloseBraceToken {
			kind = s.ReScanTemplateToken(false)
		}
		count++
		if !countOnly {
			fmt.Fprintf(out, "%s %d %d %d", strings.TrimPrefix(kind.String(), "Kind"), s.TokenStart(), s.TokenEnd(), s.TokenFlags())
			hasValue := kind == ast.KindIdentifier || kind == ast.KindPrivateIdentifier || kind >= ast.KindFirstKeyword && kind <= ast.KindLastKeyword ||
				kind == ast.KindStringLiteral || kind == ast.KindNumericLiteral || kind == ast.KindBigIntLiteral || kind == ast.KindRegularExpressionLiteral ||
				kind >= ast.KindNoSubstitutionTemplateLiteral && kind <= ast.KindTemplateTail || kind == ast.KindJsxText || kind == ast.KindJsxTextAllWhiteSpaces
			if hasValue {
				fmt.Fprintf(out, "\t%s", written(s.TokenValue()))
			}
			fmt.Fprintln(out)
		}
		if kind == ast.KindEndOfFile {
			break
		}
	}
	return count
}
func main() {
	out := bufio.NewWriterSize(os.Stdout, 65536)
	defer out.Flush()
	args := os.Args[1:]
	if args[0] != "--manifest" {
		mode := "scan"
		if len(args) > 1 {
			mode = args[1]
		}
		run(out, args[0], mode, false)
		return
	}
	manifest, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	countOnly := len(args) > 2 && args[2] == "--count"
	count, caseNumber := 0, 0
	for _, line := range strings.Split(string(manifest), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if !countOnly {
			fmt.Fprintf(out, "case %d\n", caseNumber)
		}
		count += run(out, fields[1], fields[0], countOnly)
		caseNumber++
	}
	if countOnly {
		fmt.Fprintln(out, count)
	}
}
