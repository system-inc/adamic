package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A probe for #egmxf20, not for landing: a whole open-enum value reaching a union's tag slots through an
// object literal, read by a reader that dispatches by the tag with no never read.
const zzProbeShapes = `enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius);
	return "square " + String(shape.side * 10);
}
function other(): number { return 7; }
const kind: Kind = other();
`

const zzProbeOptional = `enum Kind { Circle, Square }
interface Circle { readonly kind: Kind.Circle; readonly radius?: number }
interface Square { readonly kind: Kind.Square; readonly side: number }
type Shape = Circle | Square;
function area(shape: Shape): string {
	if (shape.kind === Kind.Circle) return "circle " + String(shape.radius ?? -1);
	return "square " + String(shape.side * 10);
}
function other(): number { return 7; }
const kind: Kind = other();
`

func TestZzProbeEnumTagRoads(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"literal", zzProbeShapes + "console.log(area({ kind, radius: 1, side: 2 }));\n"},
		{"shorthand", zzProbeShapes + "const radius = 1;\nconst side = 2;\nconsole.log(area({ kind, radius, side }));\n"},
		{"spread", zzProbeShapes + "const base = { radius: 1, side: 2 };\nconsole.log(area({ ...base, kind }));\n"},
		{"nested", zzProbeShapes + "interface Holder { readonly shape: Shape }\nconst holder: Holder = { shape: { kind, radius: 1, side: 2 } };\nconsole.log(area(holder.shape));\n"},
		{"variable", zzProbeShapes + "const made = { kind, radius: 1, side: 2 };\nconsole.log(area(made));\n"},
		{"optional", zzProbeOptional + "console.log(area({ kind, side: 2 }));\n"},
		{"control", zzProbeShapes + "console.log(area({ kind: Kind.Square, side: 2 }) + \" \" + area({ kind: Kind.Circle, radius: 1 }));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			path := filepath.Join(t.TempDir(), probe.name+extension)
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			line := func(r run) string {
				stderr := strings.ReplaceAll(strings.TrimSpace(string(r.stderr)), "\n", " ")
				if len(stderr) > 70 {
					stderr = stderr[:70]
				}
				return fmt.Sprintf("%d %q %q", r.exitCode, strings.TrimSpace(string(r.stdout)), stderr)
			}
			node := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				text := err.Error()
				if index := strings.Index(text, probe.name+extension); index >= 0 {
					text = text[index:]
				}
				if len(text) > 150 {
					text = text[:150]
				}
				t.Logf("PROBE %s%s | node %s | stops: %s", probe.name, extension, line(node), text)
				continue
			}
			sanitized, _ := nativelyUncached(t, program)
			t.Logf("PROBE %s%s | node %s | release %s | sanitized %s | js %s", probe.name, extension, line(node), line(releasedUncached(t, program)), line(sanitized), line(onJavaScriptBackend(t, program)))
		}
	}
}
