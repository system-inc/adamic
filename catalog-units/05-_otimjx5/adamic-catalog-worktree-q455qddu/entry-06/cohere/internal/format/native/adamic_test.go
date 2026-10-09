package native

import (
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
)

// adamicSources are texts whose formatting depends on the file being TypeScript and not TSX: `<T,>`'s comma,
// which Prettier keeps only where JSX could read the arrow as an element, and an angle-bracket assertion,
// which TSX does not parse at all.
var adamicSources = []string{
	"export const identity = <T,>(value: T): T => value;\n",
	"const widened = <number>value;\n",
	"interface Shape{width:number;height:number}\nexport function area(shape:Shape):number{return shape.width*shape.height}\n",
	"import type { Shape } from './geometry.a';\nexport type Maybe = Shape | undefined;\n",
}

// An Adamic `.a` file formats to the same bytes as the same text named `.ts` (#6mhafvb), and the `.tsx` control
// differs on the comma, so the name is what decides and the test could have failed.
func TestAnAdamicFileFormatsAsTypeScript(t *testing.T) {
	t.Parallel()
	formatter := Formatter{Options: formatoptions.Default()}
	if !formatter.Handles("Dog.a") {
		t.Fatal("no printer is registered for .a")
	}
	for _, source := range adamicSources {
		asTypeScript, err := formatter.Format("/repo/Dog.ts", source)
		if err != nil {
			t.Fatalf("as .ts: %v", err)
		}
		asAdamic, err := formatter.Format("/repo/Dog.a", source)
		if err != nil {
			t.Fatalf("as .a: %v", err)
		}
		if asAdamic != asTypeScript {
			t.Errorf("%q formats as .a to\n%s\nand as .ts to\n%s", source, asAdamic, asTypeScript)
		}
	}
	asTSX, err := formatter.Format("/repo/Dog.tsx", adamicSources[0])
	if err != nil {
		t.Fatal(err)
	}
	if asAdamic, _ := formatter.Format("/repo/Dog.a", adamicSources[0]); asAdamic == asTSX {
		t.Errorf("the .tsx control formats `<T,>` as .a does, so the comparison above proves nothing:\n%s", asTSX)
	}
}
