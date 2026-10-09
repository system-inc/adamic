package prettier

import (
	"strings"
	"testing"
)

// embeddedBundleCopy returns the embedded bundles as a map the test owns, so mutating it cannot
// reach the copy another parallel test is reading.
func embeddedBundleCopy(t *testing.T) map[string][]byte {
	t.Helper()

	source, err := Bundles()
	if err != nil {
		t.Fatalf("reading the embedded bundles: %v", err)
	}

	files := make(map[string][]byte, len(source.Files))
	for name, content := range source.Files {
		files[name] = append([]byte(nil), content...)
	}
	return files
}

// TestDigestDetectsOneChangedByte is the known-dirty control. A digest that has never returned a
// different value for different bytes has not been shown to detect anything, and the oracle cache is
// keyed on it, so a digest that missed a change would serve Prettier output from bundles that no
// longer exist.
//
// The byte is flipped in the last bundle, so a digest that stopped reading after the first would
// still pass a stability check and fail here.
func TestDigestDetectsOneChangedByte(t *testing.T) {
	t.Parallel()

	files := embeddedBundleCopy(t)

	before, err := DigestBundles(files)
	if err != nil {
		t.Fatalf("digesting before: %v", err)
	}

	last := BundleFiles[len(BundleFiles)-1]
	files[last][0] ^= 0x01

	after, err := DigestBundles(files)
	if err != nil {
		t.Fatalf("digesting after: %v", err)
	}

	if after == before {
		t.Fatalf("one flipped byte in %s produced the same digest, so the digest does not cover it", last)
	}
}

// TestDigestRefusesAnAbsentBundle holds the loud-failure rule: absence names the file rather than
// digesting the bundles that are there.
func TestDigestRefusesAnAbsentBundle(t *testing.T) {
	t.Parallel()

	files := embeddedBundleCopy(t)
	missing := BundleFiles[len(BundleFiles)-1]
	delete(files, missing)

	digest, err := DigestBundles(files)
	if err == nil {
		t.Fatalf("a set missing %s digested to %s instead of failing", missing, digest)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("the error does not name the absent bundle: %v", err)
	}
}

// TestDigestKeysByName covers the half of the framing the length does not.
//
// Length framing is injective over a fixed ordered list, so it catches boundary shifts on its own
// and the name looks redundant. The case the name defends is a rename: a bundle moving from
// `plugins/yaml.js` to `plugins/yml.js` with identical bytes at the same position digests
// identically under length framing alone, and BundleFiles is a hand-edited list where a rename is
// an ordinary edit.
//
// It goes through digestNamedBundles rather than re-implementing the hash, because an earlier
// version hashed inline and passed while a mutation dropping the name from the real function went
// undetected -- a test of its own arithmetic rather than of the code.
func TestDigestKeysByName(t *testing.T) {
	t.Parallel()

	content := []byte("the same bytes under either name")

	before, err := digestNamedBundles([]string{"plugins/yaml.js"}, map[string][]byte{"plugins/yaml.js": content})
	if err != nil {
		t.Fatalf("digesting under the first name: %v", err)
	}

	after, err := digestNamedBundles([]string{"plugins/yml.js"}, map[string][]byte{"plugins/yml.js": content})
	if err != nil {
		t.Fatalf("digesting under the second name: %v", err)
	}

	if before == after {
		t.Fatal("the same bytes under two different names digested identically, so the name is not in the hash")
	}
}
