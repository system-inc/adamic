package rule_testing

import (
	"encoding/json"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"testing"
)

type captureCustomLocation int

func (captureCustomLocation) MarshalJSON() ([]byte, error) { return []byte(`"anywhere"`), nil }
func TestAdamicCompiledOptionSources(t *testing.T) {
	pattern, err := esregexp.Compile("^_", "u")
	if err != nil {
		t.Fatal(err)
	}
	options := struct {
		Patterns []*esregexp.RegExp    `json:"exceptionPatterns"`
		Ignored  string                `json:"ignored,omitempty"`
		Location captureCustomLocation `json:"location"`
		Large    uint64                `json:"large"`
		Nested   map[string]any        `json:"nested"`
	}{Patterns: []*esregexp.RegExp{pattern, nil}, Large: 9007199254740993, Nested: map[string]any{"pattern": pattern}}
	got, err := marshalCapturedOptions(options)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"exceptionPatterns":["^_",null],"large":9007199254740993,"location":"anywhere","nested":{"pattern":"^_"}}`
	if string(got) != want {
		t.Fatalf("compiled option source lost or other options changed: got %s want %s", got, want)
	}
	raw, _ := json.Marshal(options)
	t.Logf("raw=%s captured=%s", raw, got)
}
