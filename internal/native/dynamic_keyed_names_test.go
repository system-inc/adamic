package native

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode"

	"github.com/system-inc/adamic/internal/ir"
)

var updateDynamicKeyedNames = flag.Bool("update-dynamic-keyed-names", false, "regenerate runtime/dynamic_keyed_names.h")

// dynamicKeyedNamesHeader is runtime/dynamic_keyed_names.h, written from internal/ir's one table of a
// run-time key read's stops and names, which the JavaScript backend emits from too.
func dynamicKeyedNamesHeader() []byte {
	var text bytes.Buffer
	text.WriteString("// dynamic_keyed_names.h: a run-time key read's stops and names (adamic_dynamic_property_keyed, in\n")
	text.WriteString("// union.c), written from internal/ir/dynamic_keyed_stops.go, the table the JavaScript backend emits\n")
	text.WriteString("// from too. Don't edit it; run go test ./internal/native -run '^TestDynamicKeyedNamesHeaderIsCurrent$'\n")
	text.WriteString("// -args -update-dynamic-keyed-names. Only union.c includes it.\n\n")
	text.WriteString("#ifndef ADAMIC_DYNAMIC_KEYED_NAMES_H\n#define ADAMIC_DYNAMIC_KEYED_NAMES_H\n\n")
	text.WriteString("enum adamic_dynamic_keyed_stop {\n")
	for _, stop := range ir.DynamicKeyedStops {
		fmt.Fprintf(&text, "\tadamic_dynamic_keyed_stop_%s,\n", snakeCase(stop.Name))
	}
	text.WriteString("};\n\n")
	text.WriteString("static const char *const adamic_dynamic_keyed_stop_messages[] = {\n")
	for _, stop := range ir.DynamicKeyedStops {
		fmt.Fprintf(&text, "\t%s,\n", cString(stop.Message))
	}
	text.WriteString("};\n\n")
	for _, list := range []struct {
		name  string
		names []string
	}{{"prototype_method", ir.DynamicKeyedPrototypeMethods}, {"array_prototype", ir.DynamicKeyedArrayPrototypeNames.Names()}, {"regexp_string_iterator", ir.DynamicKeyedRegExpStringIteratorNames.Names()}, {"stats", ir.DynamicKeyedStatsNames.Names()}, {"stats_field", ir.DynamicKeyedStatsFieldNames}, {"error", ir.DynamicKeyedErrorNames}} {
		fmt.Fprintf(&text, "static const char *const adamic_dynamic_keyed_%s_names[] = {\n", list.name)
		for _, name := range list.names {
			fmt.Fprintf(&text, "\t%s,\n", cString(name))
		}
		text.WriteString("};\n\n")
	}
	// A closure is a struct with a flexible array member, so the closures can't be one array: union.c
	// gives each of these a function(index, length) macro and defines one static closure for each.
	for _, table := range []struct {
		name      string
		functions ir.DynamicKeyedFunctions
	}{{"array_prototype", ir.DynamicKeyedArrayPrototypeNames}, {"regexp_string_iterator", ir.DynamicKeyedRegExpStringIteratorNames}, {"stats", ir.DynamicKeyedStatsNames}} {
		fmt.Fprintf(&text, "// Each of adamic_dynamic_keyed_%s_names by its index, with Node's length for it.\n", table.name)
		fmt.Fprintf(&text, "#define ADAMIC_DYNAMIC_KEYED_%s_FUNCTIONS(function) \\\n", strings.ToUpper(table.name))
		for index, function := range table.functions {
			fmt.Fprintf(&text, "\tfunction(%d, %d)", index, function.Length)
			if index < len(table.functions)-1 {
				text.WriteString(" \\")
			}
			text.WriteString("\n")
		}
		text.WriteString("\n")
	}
	text.WriteString("#endif\n")
	return text.Bytes()
}

// snakeCase is a stop's PascalCase name as its C enumerator spells it: ObjectLiteralGetter is
// object_literal_getter.
func snakeCase(name string) string {
	var text strings.Builder
	for index, character := range name {
		if unicode.IsUpper(character) {
			if index > 0 {
				text.WriteByte('_')
			}
			character = unicode.ToLower(character)
		}
		text.WriteRune(character)
	}
	return text.String()
}

// The runtime's copy of the stop table is internal/ir's, byte for byte, so a stop added or reworded
// in Go reaches both backends or this goes red. Mutant: one message changed in the header.
//
// Not parallel: the update flag rewrites the checked-in runtime/dynamic_keyed_names.h.
func TestDynamicKeyedNamesHeaderIsCurrent(t *testing.T) {
	want := dynamicKeyedNamesHeader()
	if *updateDynamicKeyedNames {
		if err := os.WriteFile("runtime/dynamic_keyed_names.h", want, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile("runtime/dynamic_keyed_names.h")
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("runtime/dynamic_keyed_names.h is stale; run with -update-dynamic-keyed-names", err)
	}
}
