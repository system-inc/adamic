package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistencyNoForInFile = "/repository/source/ConsistencyNoForIn.ts"

// Every `for...in` in ahra at HEAD on 2026-10-01, trimmed to the loop. The first four are the ones
// upstream `guard-for-in` reported; the last four are the ones it accepted as guarded, and this rule
// reports them too, because a guarded `for...in` is still the shape being retired. The fifth is the
// case that shows why: its guard is `if (key === 'parent') continue;`, which upstream counts as a
// guard while the loop still visits inherited keys.
func TestConsistencyNoForInReportsEveryForIn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"Countries.ts:883, over a constant record", []string{
			"for(const countryCode in Countries) {",
			"    const country = Countries[countryCode];",
			"    if(country && lowerCaseCountryNames.includes(country.name.toLowerCase())) {",
			"        filteredCountries.push(country);",
			"    }",
			"}",
		}},
		{"ThemeUtilities.tsx:65, carrying the keyof cast the replacement deletes", []string{
			"for(const key in overrideTheme) {",
			"    const overrideValue = overrideTheme[key as keyof typeof overrideTheme];",
			"    const baseValue = (baseTheme as Record<string, unknown>)[key];",
			"    merged[key] = overrideValue ?? baseValue;",
			"}",
		}},
		{"Object.ts:79, mergeDeep", []string{
			"for(const key in updates) {",
			"    const keyAsString = key as string;",
			"    const updateValue = updates[keyAsString];",
			"    Reflect.set(result, keyAsString, updateValue);",
			"}",
		}},
		{"Object.ts:112, getValueForKeyRecursively, returning out of the loop", []string{
			"for(const currentKey in object) {",
			"    const currentValue = object[currentKey];",
			"    if(typeof currentValue === 'object' && currentValue !== null) {",
			"        return currentValue;",
			"    }",
			"}",
		}},
		{"ReactComponentNoDestructuringRule.ts:80, a continue that upstream reads as a guard", []string{
			"for(const key in record) {",
			"    if(key === 'parent') continue;",
			"    const child = record[key];",
			"    walk(child);",
			"}",
		}},
		{"Object.ts:44, guarded with Object.hasOwn", []string{
			"for(const key in object) {",
			"    if(Object.hasOwn(object, key)) {",
			"        result[key] = object[key];",
			"    }",
			"}",
		}},
		{"useTableRowSelectionSubscription.ts:73, a counting loop with a single if", []string{
			"let count = 0;",
			"for(const key in snapshot) {",
			"    if(snapshot[key]) count++;",
			"}",
		}},
		{"ClassName.ts:29, building a string", []string{
			"let result = '';",
			"for(const key in value) {",
			"    if(value[key]) {",
			"        if(result) result += ' ';",
			"        result += key;",
			"    }",
			"}",
		}},
		{"var and let declarations, and a bare identifier", []string{
			"for(var a in object) use(a);",
			"for(let b in object) use(b);",
			"let c;",
			"for(c in object) use(c);",
		}},
		{"nested inside a for...of", []string{
			"for(const item of items) {",
			"    for(const key in item) use(key);",
			"}",
		}},
		{"an empty body is still the statement", []string{
			"for(const key in object);",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := strings.Join(testCase.lines, "\n") + "\n"
			result := rule_testing.Run(t, ConsistencyNoForIn, consistencyNoForInFile, sourceText)

			want := strings.Count(sourceText, " in ") - strings.Count(sourceText, "' in ")
			wantIds := make([]string, 0, want)
			for range want {
				wantIds = append(wantIds, consistencyNoForInId)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The finding starts at the `for` keyword, so the line a reader is sent to is the loop
			// header rather than the comment or blank line above it.
			for _, diagnostic := range result.Diagnostics {
				if !strings.HasPrefix(sourceText[diagnostic.Range.Pos():], "for(") {
					t.Fatalf("finding starts at %q, not at the for keyword", sourceText[diagnostic.Range.Pos():][:10])
				}
			}
		})
	}
}

// The shapes the replacement is written in, and the constructs that share syntax with `for...in`
// without being it.
func TestConsistencyNoForInStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"Countries.ts:883 rewritten to Object.entries", []string{
			"for(const [countryCode, country] of Object.entries(Countries)) {",
			"    if(lowerCaseCountryNames.includes(country.name.toLowerCase())) filteredCountries.push(country);",
			"    use(countryCode);",
			"}",
		}},
		{"Object.keys with for...of", []string{
			"for(const key of Object.keys(object)) use(key);",
		}},
		{"for await...of", []string{
			"async function run() {",
			"    for await(const chunk of stream) use(chunk);",
			"}",
		}},
		// A membership test, not a loop. Written inside a `for` header too, where the parser has to
		// tell the operator from the statement.
		{"the in operator", []string{
			"if('name' in object) use(object.name);",
			"const hasName = 'name' in object;",
			"for(let index = 0; ('length' in object) && index < 3; index++) use(index);",
		}},
		{"a classic for loop", []string{
			"for(let index = 0; index < items.length; index++) use(items[index]);",
		}},
		{"for...in inside a string and a comment", []string{
			"// for(const key in object) is what this file used to do",
			"const text = 'for(const key in object) {}';",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoForIn, consistencyNoForInFile, strings.Join(testCase.lines, "\n")+"\n")
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The rewrite is not meaning-preserving in general (a prototype with enumerable properties, a body
// that mutates what it walks), so the finding carries no fix.
func TestConsistencyNoForInProposesNoFix(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoForIn, consistencyNoForInFile, "for(const key in object) use(key);\n")
	rule_testing.ExpectFindings(t, result, consistencyNoForInId)
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatalf("expected no fix and no suggestion, got %d and %d",
			len(result.Diagnostics[0].Fixes), len(result.Diagnostics[0].Suggestions))
	}
}
