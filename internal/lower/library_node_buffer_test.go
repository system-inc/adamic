package lower

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNodeBufferRefusals(t *testing.T) {
	t.Parallel()
	const indexOfString = "a Buffer.indexOf string other than a literal free of lone surrogates, on which Node aborts,"
	for _, probe := range []struct{ name, source, reason string }{
		{"buffer alloc", "import { Buffer } from 'node:buffer'; Buffer.alloc(2);", "node:buffer.BufferConstructor.alloc"},
		{"buffer read", "import { Buffer } from 'node:buffer'; console.log(Buffer.from('abc').byteOffset);", "Buffer.byteOffset"},
		{"hash inherited read", "import { createHash } from 'node:crypto'; console.log(createHash('sha256').writable);", "writable"},
		{"hash copy", "import { createHash } from 'node:crypto'; createHash('sha256').copy();", "Hash.copy"},
		{"crypto random", "import { randomBytes } from 'node:crypto'; randomBytes(2);", "node:crypto.randomBytes"},
		{"crypto constructor", "import { Hash } from 'node:crypto'; const ctor = Hash;", "node:crypto.Hash"},
		{"buffer isUtf8", "import { isUtf8, Buffer } from 'node:buffer'; isUtf8(Buffer.from('abc'));", "node:buffer.isUtf8"},
		{"encoding", "import { Buffer } from 'node:buffer'; function encode(encoding: 'hex' | 'utf8'): string { return Buffer.from('abc', encoding).toString(); } console.log(encode('hex'));", "census literal"},
		{"algorithm", "import { createHash } from 'node:crypto'; createHash('sha1');", "sha256 algorithm"},
		{"digest", "import { createHash } from 'node:crypto'; createHash('sha256').digest('base64');", "outside hex"},
		{"catch", "import { createHash } from 'node:crypto'; try { createHash('sha256').update('a').digest('hex'); } catch {}", "catchable .code contract"},
		{"detached", "import { createHash } from 'node:crypto'; const hash = createHash('sha256'); const update = hash.update;", "Hash.update"},
		{"typed array view", "import { Buffer } from 'node:buffer'; const bytes: Uint8Array = Buffer.from('abc'); console.log(`${bytes.length}`);", "another object type"},
		{"structural hash view", "import { createHash } from 'node:crypto'; const hash: { digest: (encoding: 'hex') => string } = createHash('sha256'); console.log(hash.digest('hex'));", "another object type"},
		{"allocUnsafe spread", "import { Buffer } from 'node:buffer'; const sizes: [number] = [2]; console.log(`${Buffer.allocUnsafe(...sizes).length}`);", "Buffer.allocUnsafe outside allocUnsafe(size)"},
		{"slice spread", "import { Buffer } from 'node:buffer'; const range: [number, number] = [1, 2]; console.log(`${Buffer.from('abc').slice(...range).length}`);", "Buffer.slice outside slice(start, end)"},
		{"subarray spread", "import { Buffer } from 'node:buffer'; const range: [number, number] = [1, 2]; console.log(`${Buffer.from('abc').subarray(...range).length}`);", "Buffer.subarray outside subarray(start, end)"},
		{"slice of never", "import { Buffer } from 'node:buffer'; function stop(): never { throw new Error('stop'); } console.log(`${Buffer.from('abc').slice(stop()).length}`);", "a Buffer.slice offset that isn't number or undefined"},
		// ?. on a host method meets optional_host_call.go's stop first (#3625x53). nodeBufferSubarray's own
		// optional-call stop sits behind it, the guard for the day that one lifts.
		{"optional slice", "import { Buffer } from 'node:buffer'; const bytes = Buffer.from('abc'); const view = bytes?.slice(1); console.log(`${view.length}`);", "doesn't guard an optional receiver"},
		{"indexOf encoding", "import { Buffer } from 'node:buffer'; console.log(`${Buffer.from('abc').indexOf('b', 0, 'utf8')}`);", "Buffer.indexOf outside indexOf(value, byteOffset)"},
		{"indexOf encoding as offset", "import { Buffer } from 'node:buffer'; console.log(`${Buffer.from('abc').indexOf('b', 'utf8')}`);", "a Buffer.indexOf byteOffset that isn't number or undefined"},
		{"indexOf spread", "import { Buffer } from 'node:buffer'; const found: [string] = ['b']; console.log(`${Buffer.from('abc').indexOf(...found)}`);", "Buffer.indexOf outside indexOf(value, byteOffset)"},
		// A Buffer needle meets the host view stop first, a Buffer seen as indexOf's string | number | Uint8Array.
		{"indexOf of a Buffer", "import { Buffer } from 'node:buffer'; console.log(`${Buffer.from('abc').indexOf(Buffer.from('b'))}`);", "Buffer or Hash viewed as another object type"},
		{"indexOf of a Uint8Array", "import { Buffer } from 'node:buffer'; console.log(`${Buffer.from('abc').indexOf(new Uint8Array([98]))}`);", "a Buffer.indexOf value that isn't a string or a number"},
		{"indexOf of a string or a number", "import { Buffer } from 'node:buffer'; function find(value: string | number): number { return Buffer.from('abc').indexOf(value); } console.log(`${find('b')}`);", "a Buffer.indexOf value that isn't a string or a number"},
		// Node aborts on a lone surrogate needle, so only a literal proven free of one lowers: a literal holding
		// one, a string parameter and a const string read from a local, a Read and not a constant, all stop.
		{"indexOf of a lone surrogate", "import { Buffer } from 'node:buffer'; console.log(`${Buffer.from('abc').indexOf('a\\ud800')}`);", indexOfString},
		{"indexOf of a string parameter", "import { Buffer } from 'node:buffer'; function find(value: string): number { return Buffer.from('abc').indexOf(value); } console.log(`${find('b')}`);", indexOfString},
		{"indexOf of a const local", "import { Buffer } from 'node:buffer'; const newline = '\\n'; console.log(`${Buffer.from('a\\nb').indexOf(newline)}`);", indexOfString},
		{"optional indexOf", "import { Buffer } from 'node:buffer'; const bytes = Buffer.from('abc'); const found = bytes?.indexOf('b'); console.log(`${found}`);", "doesn't guard an optional receiver"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
}

// A Buffer is held as a Uint8Array natively, but the two are different objects on Node: a
// Uint8Array's toString prints its numbers and its slice copies, where a Buffer's decodes and
// shares. So no relation may see either as the other, at any depth. Each road is pinned at the
// stop it meets today, all of them ahead of sameKeeping: adamic/invariant-mutable for a writable
// container, method-signature-style for a method's parameter, and the host view registration
// (nodeHostViewAt) for every readonly or one-way road, since a Uint8Array view of a Buffer reaches
// members no registration names. sameKeeping's own order (typedArrayKind, Buffer-blind, asked
// before nodeBufferView) stands behind all of them, so answering Uint8Array for a Buffer there
// changes none of these today; with the host view stop taken out, it is what refuses the readonly
// roads. The day registration names those members, the host view pins turn red, and the relation
// pins re-read against sameKeeping. Each control holds Buffer on both sides, and lowers.
func TestNodeBufferStaysApartFromUint8Array(t *testing.T) {
	t.Parallel()
	const prefix = "import { Buffer } from 'node:buffer';\n"
	const hostView = "which reaches Buffer<ArrayBufferLike>.BYTES_PER_ELEMENT, a member no host registration names yet"
	for _, probe := range []struct{ name, source, reason string }{
		{"array", "const buffers: Buffer[] = [Buffer.from('a')];\nconst seen: Uint8Array[] = buffers;\nconsole.log(`${seen.length}`);", "Buffer<ArrayBufferLike>[] seen as Uint8Array<ArrayBufferLike>[], which can write Uint8Array<ArrayBufferLike> where Buffer<ArrayBufferLike> is read"},
		{"field", "const held: { b: Buffer } = { b: Buffer.from('a') };\nconst seen: { b: Uint8Array } = held;\nconsole.log(`${seen.b.length}`);", "{ b: Buffer<ArrayBufferLike>; } seen as { b: Uint8Array<ArrayBufferLike>; }, which can write"},
		{"readonly array", "const buffers: readonly Buffer[] = [Buffer.from('a')];\nconst seen: readonly Uint8Array[] = buffers;\nconsole.log(`${seen.length}`);", "3:37: stage 0 can't lower a value of type readonly Buffer<ArrayBufferLike>[] holding a host object seen as readonly Uint8Array<ArrayBufferLike>[], " + hostView},
		{"readonly field", "const held: { readonly b: Buffer } = { b: Buffer.from('a') };\nconst seen: { readonly b: Uint8Array } = held;\nconsole.log(`${seen.b.length}`);", "3:42: stage 0 can't lower a value of type { readonly b: Buffer<ArrayBufferLike>; } holding a host object seen as { readonly b: Uint8Array<ArrayBufferLike>; }, " + hostView},
		{"callback parameter", "const take = (bytes: Uint8Array): number => bytes.length;\nconst seen: (buffer: Buffer) => number = take;\nconsole.log(`${seen(Buffer.from('a'))}`);", "3:42: stage 0 can't lower a value of type (bytes: Uint8Array<ArrayBufferLike>) => number holding a host object seen as (buffer: Buffer<ArrayBufferLike>) => number, " + hostView},
		{"callback result", "const make = (): Buffer => Buffer.from('a');\nconst seen: () => Uint8Array = make;\nconsole.log(`${seen().length}`);", "3:32: stage 0 can't lower a value of type () => Buffer<ArrayBufferLike> holding a host object seen as () => Uint8Array<ArrayBufferLike>, " + hostView},
		{"method parameter", "const taker: { take(bytes: Buffer): number } = { take: (bytes: Buffer): number => bytes.length };\nconst seen: { take(bytes: Uint8Array): number } = taker;\nconsole.log(`${seen.take(new Uint8Array(1))}`);", "a function taking Buffer<ArrayBufferLike> seen as one taking Uint8Array<ArrayBufferLike> (tsc relates a method's parameters both ways)"},
		{"cast to Uint8Array through unknown", "const seen = Buffer.from('a') as unknown as Uint8Array;\nconsole.log(`${seen.length}`);", "a cast the runtime can't check"},
		{"cast to Buffer through unknown", "const seen = new Uint8Array(1) as unknown as Buffer;\nconsole.log(seen.toString('hex'));", "a cast the runtime can't check"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, prefix+probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q, got %v", probe.reason, err)
			}
		})
	}
	for _, control := range []struct{ name, source string }{
		{"array", "const buffers: Buffer[] = [Buffer.from('a')];\nconst seen: Buffer[] = buffers;\nconsole.log(`${seen.length}`);"},
		{"field", "const held: { b: Buffer } = { b: Buffer.from('a') };\nconst seen: { b: Buffer } = held;\nconsole.log(`${seen.b.length}`);"},
		{"readonly array", "const buffers: readonly Buffer[] = [Buffer.from('a')];\nconst seen: readonly Buffer[] = buffers;\nconsole.log(`${seen.length}`);"},
		{"readonly field", "const held: { readonly b: Buffer } = { b: Buffer.from('a') };\nconst seen: { readonly b: Buffer } = held;\nconsole.log(`${seen.b.length}`);"},
		{"callback parameter", "const take = (bytes: Buffer): number => bytes.length;\nconst seen: (buffer: Buffer) => number = take;\nconsole.log(`${seen(Buffer.from('a'))}`);"},
		{"callback result", "const make = (): Buffer => Buffer.from('a');\nconst seen: () => Buffer = make;\nconsole.log(`${seen().length}`);"},
	} {
		t.Run(control.name+" kept as Buffer", func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, prefix+control.source); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}

// Buffer.allocUnsafe(size) of a number is buffer_alloc_unsafe, which makes a Buffer, a Uint8Array as
// Buffer.from's are, and can throw: Node's RangeError for a size out of range or one it can't allocate,
// so the function holding it is one a throw can leave.
func TestNodeBufferAllocUnsafeLowers(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "import { Buffer } from 'node:buffer';\nfunction make(size: number): number { return Buffer.allocUnsafe(size).length; }\nconsole.log(`${make((16000 + 16000 / 61) | 0)}`);")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if call, isCall := node.(ir.NodeBufferCall); isCall {
				found++
				if call.Function != "buffer_alloc_unsafe" || call.Returns != ir.Uint8Array || len(call.Arguments) != 1 || call.Arguments[0].Type() != ir.Number {
					t.Errorf("Buffer.allocUnsafe lowered to %+v", call)
				}
				if !function.MayThrow {
					t.Errorf("%s holds Buffer.allocUnsafe and isn't marked as able to throw", function.Name)
				}
			}
			return true
		})
	}
	if found != 1 {
		t.Fatalf("want one buffer_alloc_unsafe in a function, found %d", found)
	}
}

// A Buffer's slice and subarray are one view sharing its bytes, as on Node: TypedArraySubarray over the
// Buffer, whose type stays Uint8Array as every Buffer's, and never a copy (buffer_copy). Each argument is
// held as number | undefined, so an explicit undefined is the default as an omitted one is.
func TestNodeBufferSliceLowers(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "import { Buffer } from 'node:buffer';\nfunction views(bytes: Buffer, end: number | undefined): number {\n\treturn bytes.slice().length + bytes.slice(1).length + bytes.subarray(1, -1).length + bytes.slice(undefined, end).length;\n}\nconsole.log(`${views(Buffer.from('abcdef'), 3)}`);")
	if err != nil {
		t.Fatal(err)
	}
	counts := []int{}
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			switch node := node.(type) {
			case ir.TypedArraySubarray:
				counts = append(counts, len(node.Arguments))
				if node.Type() != ir.Uint8Array || node.Array.Type() != ir.Uint8Array {
					t.Errorf("Buffer view lowered as %+v, not a Uint8Array", node)
				}
				for _, argument := range node.Arguments {
					if argument.Type() != ir.MaybeNumber {
						t.Errorf("Buffer view argument %+v isn't number | undefined", argument)
					}
				}
			case ir.NodeBufferCall:
				t.Errorf("Buffer view lowered through %s", node.Function)
			}
			return true
		})
	}
	if len(counts) != 4 || counts[0] != 0 || counts[1] != 1 || counts[2] != 2 || counts[3] != 2 {
		t.Fatalf("want four Buffer views of 0, 1, 2 and 2 arguments, found %v", counts)
	}
}

// Must 10: a Uint8Array's slice copies on Node, where a Buffer's shares, so it never lowers to a view.
// It stays refused while a Buffer's slice beside it is admitted, and a Uint8Array's subarray still lowers.
func TestNodeBufferSliceLeavesUint8ArraySliceRefused(t *testing.T) {
	t.Parallel()
	const prefix = "import { Buffer } from 'node:buffer';\nconst shared = Buffer.from('abc').slice(1);\nconst bytes = new Uint8Array([1, 2, 3]);\n"
	if _, err := lowerSource(t, prefix+"console.log(`${shared.length} ${bytes.slice(1).length}`);"); err == nil || !strings.Contains(err.Error(), "typed array method slice") {
		t.Fatalf("want a Uint8Array's slice refused as %q, got %v", "typed array method slice", err)
	}
	if _, err := lowerSource(t, prefix+"console.log(`${shared.length} ${bytes.subarray(1).length}`);"); err != nil {
		t.Fatalf("want a Uint8Array's subarray beside a Buffer's slice lowered, got %v", err)
	}
}

// Buffer.indexOf of a string or a number is buffer_index_of, a number that reads the Buffer and makes
// nothing: its arguments are the Buffer, the value as it lowered, and the offset as number | undefined
// when one is given, so an explicit undefined is the default as an omitted one is. Each string literal,
// '\n' and '>' among them, arrives as the ir.StringConstant nodeBufferIndexOf proves free of lone
// surrogates.
func TestNodeBufferIndexOfLowers(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "import { Buffer } from 'node:buffer';\nfunction find(bytes: Buffer, from: number | undefined): number {\n\treturn bytes.indexOf('\\n') + bytes.indexOf(62, 1) + bytes.indexOf('AC', from) + bytes.indexOf('>', undefined);\n}\nconsole.log(`${find(Buffer.from('>A\\nAC'), -2)}`);")
	if err != nil {
		t.Fatal(err)
	}
	found := []string{}
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if call, isCall := node.(ir.NodeBufferCall); isCall {
				if call.Function != "buffer_index_of" || call.Returns != ir.Number || len(call.Arguments) < 2 || call.Arguments[0].Type() != ir.Uint8Array {
					t.Errorf("Buffer.indexOf lowered to %+v", call)
					return true
				}
				shape := "number"
				if constant, isConstant := call.Arguments[1].(ir.StringConstant); isConstant {
					shape = strconv.Quote(program.Strings[constant.Index])
				} else if call.Arguments[1].Type() != ir.Number {
					t.Errorf("Buffer.indexOf value lowered as %+v, neither a string constant nor a number", call.Arguments[1])
				}
				for _, offset := range call.Arguments[2:] {
					shape += " offset"
					if offset.Type() != ir.MaybeNumber {
						t.Errorf("Buffer.indexOf offset %+v isn't number | undefined", offset)
					}
				}
				found = append(found, shape)
			}
			return true
		})
	}
	want := []string{`"\n"`, "number offset", `"AC" offset`, `">" offset`}
	slices.Sort(found)
	slices.Sort(want)
	if !slices.Equal(found, want) {
		t.Fatalf("want buffer_index_of of %v, found %v", want, found)
	}
}

// Must 10: a Uint8Array's indexOf finds one element value and never a string, so it never reaches
// buffer_index_of. It stays refused as typedArrayExpression refuses it while a Buffer's indexOf beside it
// is admitted.
func TestNodeBufferIndexOfLeavesUint8ArrayIndexOfRefused(t *testing.T) {
	t.Parallel()
	const prefix = "import { Buffer } from 'node:buffer';\nconst found = Buffer.from('abc').indexOf('b');\nconst bytes = new Uint8Array([1, 2, 3]);\n"
	if _, err := lowerSource(t, prefix+"console.log(`${found} ${bytes.indexOf(2)}`);"); err == nil || !strings.Contains(err.Error(), "typed array method indexOf") {
		t.Fatalf("want a Uint8Array's indexOf refused as %q, got %v", "typed array method indexOf", err)
	}
	if _, err := lowerSource(t, prefix+"console.log(`${found} ${bytes.length}`);"); err != nil {
		t.Fatalf("want a Buffer's indexOf beside a Uint8Array lowered, got %v", err)
	}
}
