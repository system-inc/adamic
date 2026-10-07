package checker

import (
	"strings"
	"testing"
)

func TestWave07RegexStructure(t *testing.T) {
	p := new(Program)
	for _, input := range []string{"wave07-regex-structure\ng\na/b", "wave07-regex-structure\nu\n世界/", "wave07-regex-structure\n\n["} {
		text, err := p.wave07RegexStructure(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		values := decodedFields(t, text)
		if len(values) < 4 || values[0] != "1" || values[1] != "wave07-regex-structure" {
			t.Fatal(values)
		}
		if strings.HasSuffix(input, "[") && values[2] != "0" {
			t.Fatal("malformed class was scanned", values)
		}
		if strings.HasSuffix(input, "a/b") && (values[2] != "1" || values[3] != "3") {
			t.Fatal("characters omitted", values)
		}
	}
	for _, input := range []string{"wave07-regex-structure", "wave07-regex-structure\ng"} {
		if _, err := p.wave07RegexStructure(nil, input); err == nil {
			t.Fatal("missing field accepted")
		}
	}
}
