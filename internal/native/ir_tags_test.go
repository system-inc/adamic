package native

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native_runtime"
)

func TestIRTypeTags(t *testing.T) {
	t.Parallel()
	contents, err := native_runtime.Sources.ReadFile("c/adamic.h")
	if err != nil {
		t.Fatal(err)
	}
	body := regexp.MustCompile(`(?s)enum adamic_ir_type\s*\{([^}]+)\};`).FindStringSubmatch(string(contents))
	if body == nil {
		t.Fatal("missing C IR enum")
	}
	var definitions [][]string
	for _, line := range strings.Split(strings.TrimSpace(body[1]), "\n") {
		definition := regexp.MustCompile(`^\s*adamic_ir_([A-Za-z0-9]+)\s*=\s*([0-9]+),\s*$`).FindStringSubmatch(line)
		if definition == nil {
			t.Fatalf("unrecognized C IR declaration %q", line)
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
			t.Fatalf("unknown or duplicate C IR tag %s", name)
		}
		seen[name] = true
		if int(tag) != number {
			t.Errorf("%s: C=%d Go=%d", name, number, tag)
		}
	}
	for name := range tags {
		if !seen[name] {
			t.Errorf("missing C IR tag %s", name)
		}
	}
}
