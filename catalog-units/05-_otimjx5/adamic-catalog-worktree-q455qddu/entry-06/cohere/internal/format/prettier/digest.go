package prettier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// DigestBundles hashes a set of bundles, so a cache of what they produced can say which bytes it
// was produced by.
//
// It lives beside the bundles because the only thing left that needs it is the test oracle. The
// corpus and differential suites key their cache of Prettier output on this digest, so rebuilding
// the fork invalidates every cached answer instead of comparing the native printers against output
// from bundles nobody loads any more. It used to be release code, stamping the bundles a shipped
// binary carried, and no shipped binary carries them now.
func DigestBundles(files map[string][]byte) (string, error) {
	return digestNamedBundles(BundleFiles, files)
}

// digestNamedBundles is the hash itself, over a caller-supplied name list.
//
// The list is a parameter rather than a reference to BundleFiles so the framing can be tested. With
// the names hardcoded, no test could vary them, and a mutation dropping the name from the hash
// passed the whole package -- the property was untestable through the exported function and so was
// unverified despite having a test named for it.
//
// Each bundle is preceded by its name and its length, and only the name is load-bearing.
//
// The name defends a rename: the same bytes moving from `plugins/yaml.js` to `plugins/yml.js` at
// the same position digest identically without it, and BundleFiles is a hand-edited list where a
// rename is an ordinary edit. TestDigestKeysByName is its sole detector, confirmed by mutation.
//
// The length defends a boundary shift -- "AB" then "C" concatenating identically to "A" then "BC" --
// but only when the names cannot already separate the halves, and in this list they always can,
// because the names are distinct. So it is redundant here rather than load-bearing, and it is kept
// because it costs nothing and stops being redundant the moment two entries could share a name. It
// is not covered by a test, deliberately: a test for it would have to construct a name collision
// this list cannot contain, which would assert a property of its own fixture.
func digestNamedBundles(names []string, files map[string][]byte) (string, error) {
	hash := sha256.New()

	for _, name := range names {
		content, present := files[name]
		if !present {
			return "", fmt.Errorf("the Prettier bundle %s is absent, so these bytes cannot be digested", name)
		}

		fmt.Fprintf(hash, "%s\x00%d\x00", name, len(content))
		hash.Write(content)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
