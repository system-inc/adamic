package program

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// A key is made only over bytes the reader could have seen: a present path whose size or time moved between
// the read and the hashing is declined, since hashing its new bytes would key findings on a stylesheet no rule
// read. And a key that is made holds until anything it covers changes, an absence included.
func TestDesignSystemKeyDeclinesAMovedFileAndHoldsUntilAChange(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	theme := filepath.Join(directory, "theme.css")
	missing := filepath.Join(directory, "missing.css")
	if err := os.WriteFile(theme, []byte("@theme {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(theme)
	if err != nil {
		t.Fatal(err)
	}
	read := rule.FileRead{Path: theme, Present: true, Size: info.Size(), ModifiedNanoseconds: info.ModTime().UnixNano()}
	absent := rule.FileRead{Path: missing}

	key, made := designSystemKeyFromReads([]rule.FileRead{read, absent}, PathAnchor{})
	if !made || !key.stillHolds(PathAnchor{}) {
		t.Fatalf("a key over unmoved reads was not made, or does not hold: made %v", made)
	}

	moved := read
	moved.ModifiedNanoseconds -= int64(time.Second)
	if _, made := designSystemKeyFromReads([]rule.FileRead{moved}, PathAnchor{}); made {
		t.Error("a file whose time moved after it was read was keyed anyway")
	}
	grown := read
	grown.Size++
	if _, made := designSystemKeyFromReads([]rule.FileRead{grown}, PathAnchor{}); made {
		t.Error("a file whose size moved after it was read was keyed anyway")
	}

	if err := os.WriteFile(theme, []byte("@theme { --color-x: red; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if key.stillHolds(PathAnchor{}) {
		t.Error("the key still holds after the stylesheet it covers was edited")
	}
	if err := os.WriteFile(theme, []byte("@theme {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !key.stillHolds(PathAnchor{}) {
		t.Error("the key does not hold once the stylesheet is back to the bytes it covers")
	}
	if err := os.WriteFile(missing, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if key.stillHolds(PathAnchor{}) {
		t.Error("the key still holds after a path it recorded absent was created")
	}

	var unset *DesignSystemKey
	if unset.stillHolds(PathAnchor{}) || (&DesignSystemKey{}).stillHolds(PathAnchor{}) {
		t.Error("a missing or zero key holds, which would replay findings produced under no design system")
	}
}
