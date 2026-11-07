package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Two settles of one key, as two runs of the same reads racing to place it, keep the first copy and drop the second;
// when the second's bytes differ, one settle-difference line names the key and the first differing file, and when
// they're alike nothing is logged. Either way the kept copy and where placeSettled says it is stay as they were
// (#dc483fc). Mutant: placeSettled skipping the comparison: no line, and this test fails.
// Not parallel: test.Setenv.
func TestASecondCopyOfAKeyIsComparedWithTheKeptOne(test *testing.T) {
	for _, check := range []struct {
		name, second, want string
	}{
		{"different", "rebuilt elsewhere", "settle-difference %s product at byte 0\n"},
		{"alike", "built", ""},
	} {
		test.Run(check.name, func(test *testing.T) {
			cache := test.TempDir()
			log := filepath.Join(test.TempDir(), "builds.log")
			test.Setenv("ADAMIC_BUILD_LOG", log)
			key := strings.Repeat("cd", 32)
			first := filepath.Join(cache, "pending", "trace-one", "name-key")
			second := filepath.Join(cache, "pending", "trace-two", "name-key")
			write(test, first, "product", "built")
			write(test, second, "product", check.second)
			product := filepath.Join(cache, key)
			for _, settled := range []string{first, second} {
				if placed := placeSettled(
					cache,
					key,
					journalEntry{NameKey: "name-key", Directory: settled},
				); placed != product {
					test.Fatalf("placeSettled returned %s, not the product at %s", placed, product)
				}
			}
			if content, readError := os.ReadFile(
				filepath.Join(product, "product"),
			); readError != nil || string(content) != "built" {
				test.Errorf("the kept copy reads %q, %v, want the first's", content, readError)
			}
			if _, statError := os.Stat(second); !os.IsNotExist(statError) {
				test.Errorf("the second copy stayed at %s: %v", second, statError)
			}
			lines, _ := os.ReadFile(log)
			want := check.want
			if want != "" {
				want = strings.Replace(want, "%s", key, 1)
			}
			if string(lines) != want {
				test.Errorf("logged %q, want %q", lines, want)
			}
		})
	}
}
