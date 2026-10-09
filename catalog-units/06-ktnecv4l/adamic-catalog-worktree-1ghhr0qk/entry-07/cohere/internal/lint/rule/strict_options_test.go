package rule

import (
	"encoding/json"
	"strings"
	"testing"
)

type strictInner struct {
	Pattern string `json:"pattern"`
}

type strictEmbedded struct {
	Shared bool `json:"shared"`
}

type strictCustom struct{ value string }

func (custom *strictCustom) UnmarshalJSON(raw []byte) error {
	custom.value = string(raw)
	return nil
}

type strictOptions struct {
	strictEmbedded
	EnforceForTSTypes bool          `json:"enforceForTSTypes"`
	Paths             []strictInner `json:"paths"`
	Ignore            []string      // untagged: reached through Go's case-insensitive match
	Custom            strictCustom  `json:"custom"`
}

// TestUnmarshalOptionsRefusesWhatUpstreamRefuses is #4n972g9: an unknown key, and a key spelled in
// a different case from its tag, are refused rather than dropped. The baseline case comes first, so
// a refusal below cannot be explained by the decoder refusing everything.
func TestUnmarshalOptionsRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()
	var decoded strictOptions
	valid := `{"shared": true, "enforceForTSTypes": true, "paths": [{"pattern": "x"}], "ignore": ["a"], "custom": {"Anything": 1}}`
	if err := UnmarshalOptions([]byte(valid), &decoded); err != nil {
		t.Fatalf("baseline: a valid config was refused: %v", err)
	}
	if !decoded.Shared || !decoded.EnforceForTSTypes || len(decoded.Paths) != 1 || decoded.Paths[0].Pattern != "x" ||
		len(decoded.Ignore) != 1 || decoded.Custom.value == "" {
		t.Fatalf("baseline: decoded %+v", decoded)
	}

	refused := []struct {
		name    string
		raw     string
		mention string
	}{
		{"unknown key at the top", `{"ignorePatern": ["a"]}`, "ignorePatern"},
		{"unknown key inside a list element", `{"paths": [{"pattern": "x", "message": "y"}]}`, "message"},
		{"tagged key in the wrong case", `{"enforceForTsTypes": true}`, `"enforceForTsTypes" is spelled "enforceForTSTypes"`},
		{"tagged key in the wrong case inside a list", `{"paths": [{"Pattern": "x"}]}`, `"paths.0.Pattern" is spelled "paths.0.pattern"`},
		{"promoted key in the wrong case", `{"Shared": true}`, `"Shared" is spelled "shared"`},
	}
	for _, testCase := range refused {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var target strictOptions
			err := UnmarshalOptions([]byte(testCase.raw), &target)
			if err == nil || !strings.Contains(err.Error(), testCase.mention) {
				t.Fatalf("want a refusal naming %q, got %v", testCase.mention, err)
			}
		})
	}

	// The other direction, each a case plain json.Unmarshal also accepts: an untagged field reached
	// in any case, and a custom unmarshaler's contents, which are its own business.
	accepted := []string{`{"Ignore": ["a"]}`, `{"IGNORE": ["a"]}`, `{"custom": {"whatever": true}}`}
	for _, raw := range accepted {
		var target strictOptions
		if err := UnmarshalOptions([]byte(raw), &target); err != nil {
			t.Errorf("%s was refused: %v", raw, err)
		}
		var plain strictOptions
		if err := json.Unmarshal([]byte(raw), &plain); err != nil {
			t.Errorf("control: plain json refused %s: %v", raw, err)
		}
	}
}

// TestDecodeOptionsIntoIsStrict pins that the decoder 112 rules register through uses the strict
// path, so a refusal reaches the config layer with the type named.
func TestDecodeOptionsIntoIsStrict(t *testing.T) {
	t.Parallel()
	decode := DecodeOptionsInto[strictOptions]()
	if _, err := decode([]byte(`{"enforceForTSTypes": true}`)); err != nil {
		t.Fatalf("baseline: %v", err)
	}
	_, err := decode([]byte(`{"enforceForTsTypes": true}`))
	if err == nil || !strings.Contains(err.Error(), "rule.strictOptions") {
		t.Fatalf("want a refusal naming the options type, got %v", err)
	}
}
