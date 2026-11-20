package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNodeBufferRefusals(t *testing.T) {
	t.Parallel()
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
