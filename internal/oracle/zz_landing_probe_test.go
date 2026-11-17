package oracle

// A probe for #r11wppy's operand-landing rule, in a proof commit only, never landed: a value returned
// through nested conditionals into a declared return type whose field is wider and writable. Each probe
// prints whether it lowers and, if it does, what Node, the sanitized native build and the JavaScript
// backend print.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestZZLandingProbe(t *testing.T) {
	probes := map[string]string{
		"nested_return": `interface FileInfo { readonly version: string; signature: string | undefined; }
type Encoded = Omit<FileInfo, 'signature'> & { signature?: string | false };
type Out = string | Encoded;
const state: FileInfo = { version: 'v', signature: 'v2' };
function encode(value: FileInfo, same: boolean, kept: boolean): Out {
	return same ? value.version : kept ? value : { version: value.version, signature: false };
}
const out = encode(state, false, true);
if (typeof out !== 'string') {
	out.signature = false;
}
const signature = state.signature;
console.log(signature === undefined ? 'undefined' : signature.toUpperCase());
`,
		"direct_return": `interface FileInfo { readonly version: string; signature: string | undefined; }
type Encoded = Omit<FileInfo, 'signature'> & { signature?: string | false };
const state: FileInfo = { version: 'v', signature: 'v2' };
function encode(value: FileInfo): Encoded {
	return value;
}
const out = encode(state);
out.signature = false;
const signature = state.signature;
console.log(signature === undefined ? 'undefined' : signature.toUpperCase());
`,
		"enum_literal_slot": `enum E { A = 0, B = 1, C = 2 }
const holder: { k: E } = { k: E.C };
const open: E = holder.k;
const k: E.B | E.C = open || E.B;
console.log(String(k));
`,
		"enum_literal_argument": `enum E { A = 0, B = 1, C = 2 }
const holder: { k: E } = { k: E.C };
const open: E = holder.k;
function f(k: E.B | E.C): void {
	console.log(String(k));
}
f(open || E.B);
`,
		"enum_open_slot_control": `enum E { A = 0, B = 1, C = 2 }
const holder: { k: E } = { k: E.C };
const open: E = holder.k;
const k: E = open || E.B;
console.log(String(k));
`,
		"spread_array_or_set": `const flag = Date.now() > 0;
const list: number[] = [1, 2];
const set: Set<number> = new Set([3]);
const all = [...(flag ? list : set)];
console.log(String(all.length));
`,
		"forof_array_or_set": `const flag = Date.now() > 0;
const list: number[] = [1, 2];
const set: Set<number> = new Set([3]);
for (const value of flag ? list : set) {
	console.log(String(value));
}
`,
	}
	for name, source := range probes {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name+".ts")
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			program, err := lowered(t, path)
			if err != nil {
				fmt.Printf("PROBE %s: stops: %v\n", name, err)
				return
			}
			sanitized, _ := nativelyUncached(t, program)
			javascript := onJavaScriptBackend(t, program)
			fmt.Printf("PROBE %s: LOWERS; Node %q exit %d stderr %.200q; sanitized %q exit %d stderr %.300q; javascript %q exit %d\n", name,
				truth.stdout, truth.exitCode, truth.stderr, sanitized.stdout, sanitized.exitCode, sanitized.stderr, javascript.stdout, javascript.exitCode)
		})
	}
}
