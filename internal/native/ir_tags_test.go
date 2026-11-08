package native

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestIrTypeTags(test *testing.T) {
	test.Parallel()
	contents, readFileError := runtime.ReadFile("runtime/adamic.h")
	if readFileError != nil {
		test.Fatal(readFileError)
	}
	body := regexp.MustCompile(`(?s)enum adamic_ir_type\s*\{([^}]+)\};`).FindStringSubmatch(string(contents))
	if body == nil {
		test.Fatal("missing C IR enum")
	}
	var definitions [][]string
	for _, line := range strings.Split(strings.TrimSpace(body[1]), "\n") {
		definition := regexp.MustCompile(`^\s*adamic_ir_([A-Za-z0-9]+)\s*=\s*([0-9]+),\s*$`).FindStringSubmatch(line)
		if definition == nil {
			test.Fatalf("unrecognized C IR declaration %q", line)
		}
		definitions = append(definitions, definition)
	}
	tags := ir.TypeTags()
	seen := map[string]bool{}
	for _, definition := range definitions {
		name := definition[1]
		number, _ := strconv.Atoi(definition[2])
		tag, exists := tags[name]
		if !exists || seen[name] {
			test.Fatalf("unknown or duplicate C IR tag %s", name)
		}
		seen[name] = true
		if int(tag) != number {
			test.Errorf("%s: C=%d Go=%d", name, number, tag)
		}
	}
	for name := range tags {
		if !seen[name] {
			test.Errorf("missing C IR tag %s", name)
		}
	}
}

// The sources and flags stay identical: the Go representation ABI itself must be
// a key input for both local archives and portable supplied runtime products.
func TestRuntimeKeyIncludesTypeTags(test *testing.T) {
	test.Parallel()
	want := "ir-tags " + ir.TypeTagFingerprint()
	for name, parts := range map[string][]string{
		"local":    runtimeKeyParts(nil, nil, "clang", "version", ""),
		"portable": portableKeyParts(nil, nil, "clang", ""),
	} {
		found := false
		for _, part := range parts {
			found = found || part == want
		}
		if !found {
			test.Errorf("%s cache omits representation fingerprint", name)
		}
		changed := append([]string(nil), parts...)
		for index := range changed {
			if changed[index] == want {
				changed[index] += " changed"
			}
		}
		if partsKey(parts) == partsKey(changed) {
			test.Errorf("%s cache does not distinguish changed tags", name)
		}
	}
}
