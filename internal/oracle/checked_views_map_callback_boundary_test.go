package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Storage adapters do not extend the callback ABI. Pin each existing slotless
// callback refusal to its call site; companion loops exercise storage separately.
var mapCallbackRefusalSites = map[string]string{
	"entry-convert-key-boxed":                  "12:1",
	"entry-convert-key-boxed-boolean":          "13:1",
	"entry-convert-key-boxed-optional-boolean": "13:1",
	"entry-convert-key-boxed-optional-number":  "13:1",
	"entry-convert-key-boxed-string":           "13:1",
	"entry-convert-key-packed-boolean":         "13:1",

	"entry-array-both":                       "10:2",
	"entry-array-null":                       "9:2",
	"entry-boolean-both":                     "10:2",
	"entry-boolean-null":                     "9:2",
	"entry-convert-array-null":               "10:1",
	"entry-convert-boolean-iterator":         "10:1",
	"entry-convert-boolean-null":             "10:1",
	"entry-convert-boolean-undefined":        "10:1",
	"entry-convert-callable-null":            "10:1",
	"entry-convert-false-null":               "10:1",
	"entry-convert-false-undefined":          "10:1",
	"entry-convert-iterator":                 "10:1",
	"entry-convert-maybe-boolean":            "10:1",
	"entry-convert-maybe-number":             "10:1",
	"entry-convert-nan-null":                 "10:1",
	"entry-convert-number-null":              "10:1",
	"entry-convert-object-null":              "10:1",
	"entry-convert-optional-array":           "12:1",
	"entry-convert-optional-array-schema":    "12:1",
	"entry-convert-optional-callable":        "12:1",
	"entry-convert-optional-callable-schema": "12:1",
	"entry-convert-optional-object":          "12:1",
	"entry-convert-optional-object-schema":   "12:1",
	"entry-convert-optional-string":          "12:1",
	"entry-convert-optional-string-schema":   "12:1",
	"entry-convert-owned-callback":           "6:69",
	"entry-convert-string-null":              "10:1",
	"entry-convert-tuple-null":               "10:1",
	"entry-live-mutation":                    "5:69",
	"entry-mixed-both":                       "10:2",
	"entry-mixed-null":                       "9:2",
	"entry-nominal-aggregate-both":           "6:33",
	"entry-nominal-aggregate-null":           "6:33",
	"entry-nominal-array-aggregate-both":     "5:33",
	"entry-nominal-array-aggregate-null":     "5:33",
	"entry-number-both":                      "10:2",
	"entry-number-null":                      "9:2",
	"entry-object-both":                      "10:2",
	"entry-object-null":                      "9:2",
	"entry-string-both":                      "10:2",
	"entry-string-null":                      "9:2",
}

func mapCallbackRefusal(t *testing.T, name string) {
	t.Helper()
	site, known := mapCallbackRefusalSites[name]
	if !known {
		t.Fatalf("unlisted callback boundary: %s", name)
	}
	path, pathErr := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps", name+".a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node callback control: %#v", truth)
	}
	_, err := lowered(t, path)
	message := "a Map forEach callback with a slotless value representation"
	if strings.HasPrefix(name, "entry-convert-key-") {
		message = "a Map forEach callback with an unsupported key representation"
	}
	if err == nil || !strings.Contains(err.Error(), name+".a:"+site+":") || !strings.Contains(err.Error(), message) {
		t.Fatalf("callback refusal lost or moved: %v", err)
	}
	t.Logf("Node stdout=%q; preserved callback refusal: %v", truth.stdout, err)
}

func mapStorageFixture(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	if _, boundary := mapCallbackRefusalSites[name]; boundary {
		mapCallbackRefusal(t, name)
		name += "-iteration"
	}
	return interfaceFixture(t, "nullish/maps/"+name)
}

func TestCheckedViewMapCallbackRefusals(t *testing.T) {
	names := make([]string, 0, len(mapCallbackRefusalSites))
	for name := range mapCallbackRefusalSites {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) { mapCallbackRefusal(t, name) })
	}
}
