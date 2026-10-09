package prettier

import "testing"

// The comparison engine formats an Adamic `.a` file to the same bytes as the same text named `.ts` (#6mhafvb):
// it parses with typescript, and Prettier is handed the `.ts` name, since its /\.ts$/ would otherwise keep the
// `<T,>` comma as it does for `.tsx`, which the control shows it does.
func TestTheEngineFormatsAnAdamicFileAsTypeScript(t *testing.T) {
	t.Parallel()
	engine := newTestEngine(t)
	if !engine.Handles("Dog.a") {
		t.Fatal("the engine does not handle .a")
	}
	sources := []string{
		"export const identity = <T,>(value: T): T => value;\n",
		"const widened = <number>value;\n",
		"interface Shape{width:number;height:number}\nexport function area(shape:Shape):number{return shape.width*shape.height}\n",
	}
	for _, source := range sources {
		asTypeScript, err := engine.Format("/repo/Dog.ts", source)
		if err != nil {
			t.Fatalf("as .ts: %v", err)
		}
		asAdamic, err := engine.Format("/repo/Dog.a", source)
		if err != nil {
			t.Fatalf("as .a: %v", err)
		}
		if asAdamic != asTypeScript {
			t.Errorf("%q formats as .a to\n%s\nand as .ts to\n%s", source, asAdamic, asTypeScript)
		}
	}
	asTSX, err := engine.Format("/repo/Dog.tsx", sources[0])
	if err != nil {
		t.Fatal(err)
	}
	if asAdamic, _ := engine.Format("/repo/Dog.a", sources[0]); asAdamic == asTSX {
		t.Errorf("the .tsx control formats `<T,>` as .a does, so the comparison above proves nothing:\n%s", asTSX)
	}
}
