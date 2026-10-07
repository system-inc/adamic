package wave13more

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type listenerDescriptor struct {
	Name  string `json:"name"`
	Kinds []int  `json:"kinds"`
}

var listenerPublicNames = map[string]string{
	"no_unassigned_vars.a":                             "no-unassigned-vars",
	"preserve_caught_error.a":                          "preserve-caught-error",
	"consistent_type_exports.a":                        "@typescript-eslint/consistent-type-exports",
	"wave_13_next/no_process_exit_after_output.a":      "nexus/correctness-no-process-exit-after-output",
	"wave_13_next/no_uncleared_race_timeout.a":         "nexus/correctness-no-uncleared-race-timeout",
	"wave_13_next/require_blocking_standard_streams.a": "nexus/correctness-require-blocking-standard-streams",
	"wave_13_more/no_obj_calls.a":                      "no-obj-calls",
	"wave_13_more/no_object_constructor.a":             "no-object-constructor",
	"wave_13_more/no_promise_executor_return.a":        "no-promise-executor-return",
}

func checkListenerJSON(data []byte, name string, wanted []int) error {
	var descriptor listenerDescriptor
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&descriptor); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("descriptor has trailing data: %v", err)
	}
	if descriptor.Name != name {
		return fmt.Errorf("descriptor name %q, want %q", descriptor.Name, name)
	}
	slices.Sort(descriptor.Kinds)
	if !slices.Equal(descriptor.Kinds, wanted) {
		return fmt.Errorf("descriptor kinds %v, production wants %v", descriptor.Kinds, wanted)
	}
	return nil
}

func TestWave13ListenerJSON(t *testing.T) {
	for _, item := range listenerCases {
		t.Run(item.native, func(t *testing.T) {
			name, ok := listenerPublicNames[item.native]
			if !ok {
				t.Fatal("missing public rule name")
			}
			parts := strings.Split(name, "/")
			data, err := os.ReadFile(filepath.Join("listeners", parts[len(parts)-1], "rule.json"))
			if err != nil {
				t.Fatal(err)
			}
			wanted := productionListenerKinds(t, filepath.Join("..", "..", "..", "..", "cohere", "internal", "lint", "rules", item.production))
			if err := checkListenerJSON(data, name, wanted); err != nil {
				t.Fatal(err)
			}
			t.Logf("rule.json kinds %v agree with production Go", wanted)
			var descriptor listenerDescriptor
			if err := json.Unmarshal(data, &descriptor); err != nil {
				t.Fatal(err)
			}
			descriptor.Kinds[0]++
			mutant, err := json.Marshal(descriptor)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkListenerJSON(mutant, name, wanted); err == nil {
				t.Fatal("wrong-kind descriptor mutant survived")
			} else {
				t.Logf("descriptor numeric mutant caught: %v", err)
			}
		})
	}
}
