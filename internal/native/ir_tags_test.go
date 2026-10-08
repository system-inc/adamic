package native

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestIRTypeTags(t *testing.T) {
	contents, err := runtime.ReadFile("runtime/adamic.h")
	if err != nil {
		t.Fatal(err)
	}
	definitions := regexp.MustCompile("adamic_ir_([A-Za-z0-9]+) = ([0-9]+),").FindAllStringSubmatch(string(contents), -1)
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
