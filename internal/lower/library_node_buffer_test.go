package lower

import (
	"strings"
	"testing"
)

func TestNodeBufferRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"ambient SHA direct call", "declare function createSHA256Hash(data: string): string; console.log(createSHA256Hash('x'));", "createSHA256Hash"},
		{"ambient SHA present arm", "declare function createSHA256Hash(data: string): string; function fallback(data: string): string { return data; } const crypto = true; const hash = crypto ? createSHA256Hash : fallback; console.log(hash('x'));", "createSHA256Hash"},
		{"ambient SHA mutable guard", "declare function createSHA256Hash(data: string): string; function fallback(data: string): string { return data; } let crypto = false; const hash = crypto ? createSHA256Hash : fallback; console.log(hash('x'));", "createSHA256Hash"},
		{"namespace alias", "import * as crypto from 'node:crypto'; const alias = crypto; console.log(typeof alias);", "node:crypto namespace read"},
		{"namespace reflection", "import * as crypto from 'node:crypto'; console.log(Object.keys(crypto).length);", "node:crypto namespace read"},
		{"namespace missing member", "import * as crypto from 'node:crypto'; crypto.randomBytes(2);", "node:crypto.randomBytes"},
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
