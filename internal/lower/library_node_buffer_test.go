package lower

import (
	"strings"
	"testing"
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
// shares. So the relation keeps them apart at every depth, not only at the top, and these stop
// where either is seen as the other. typedArrayKind stays Buffer-blind for that (sameKeeping
// asks it before nodeBufferView); answering Uint8Array for a Buffer there lets the nested ones
// through. Each control holds the same shape as Buffer on both sides, and lowers.
func TestNodeBufferStaysApartFromUint8Array(t *testing.T) {
	t.Parallel()
	const prefix = "import { Buffer } from 'node:buffer';\n"
	for _, probe := range []struct{ name, source, reason string }{
		{"array", "const buffers: Buffer[] = [Buffer.from('a')];\nconst seen: Uint8Array[] = buffers;\nconsole.log(`${seen.length}`);", "seen as a Uint8Array"},
		{"field", "const held: { b: Buffer } = { b: Buffer.from('a') };\nconst seen: { b: Uint8Array } = held;\nconsole.log(`${seen.b.length}`);", "seen as a { b: Uint8Array"},
		{"callback parameter", "const take = (bytes: Uint8Array): number => bytes.length;\nconst seen: (buffer: Buffer) => number = take;\nconsole.log(`${seen(Buffer.from('a'))}`);", "seen as a (buffer: Buffer"},
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
		{"callback parameter", "const take = (bytes: Buffer): number => bytes.length;\nconst seen: (buffer: Buffer) => number = take;\nconsole.log(`${seen(Buffer.from('a'))}`);"},
	} {
		t.Run(control.name+" kept as Buffer", func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, prefix+control.source); err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
		})
	}
}
