package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{
		"regex_coverage_constructors.a",
		"regex_coverage_controls.a",
		"regex_coverage_exec.a",
		"regex_coverage_flags.a",
		"regex_coverage_methods.a",
		"regex_coverage_replace.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
	for _, name := range regexCoverageRefusals {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/regex_coverage_refused/" + name.file + ".a", false, false})
	}
}

var regexCoverageRefusals = []struct {
	file    string
	message string
}{
	{"singleton_literal", `RegExp refused: V8 does not case-fold singleton \q class strings under iv departs from ECMA-262 22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString; native and JavaScript must agree`},
	{"singleton_constructor", `RegExp refused: V8 does not case-fold singleton \q class strings under iv departs from ECMA-262 22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString; native and JavaScript must agree`},
	{"singleton_intersection", `RegExp refused: V8 does not case-fold singleton \q class strings under iv departs from ECMA-262 22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString; native and JavaScript must agree`},
	{"singleton_subtraction", `RegExp refused: V8 does not case-fold singleton \q class strings under iv departs from ECMA-262 22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString; native and JavaScript must agree`},
	{"singleton_negated", `RegExp refused: V8 does not case-fold singleton \q class strings under iv departs from ECMA-262 22.2.2.9 CompileToCharSet and 22.2.2.10 CompileClassSetString; native and JavaScript must agree`},
	{"modifier_enable", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_ascii_property", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_class_string", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_disable", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_word", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_nonword", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_word_class", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_negated_class", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_property", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_alternative", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_quantified", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"modifier_lookbehind", `RegExp refused: V8 leaks or drops a scoped i modifier in a later Unicode class or word escape departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.2.7.4 UpdateModifiers; native and JavaScript must agree`},
	{"mixed_literal", `RegExp refused: V8 can repeat an earlier empty match and hang replacement for mixed-length \q alternatives departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex; native and JavaScript must agree`},
	{"mixed_constructor", `RegExp refused: V8 can repeat an earlier empty match and hang replacement for mixed-length \q alternatives departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex; native and JavaScript must agree`},
	{"mixed_dead", `RegExp refused: V8 can repeat an earlier empty match and hang replacement for mixed-length \q alternatives departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex; native and JavaScript must agree`},
	{"mixed_union", `RegExp refused: V8 can repeat an earlier empty match and hang replacement for mixed-length \q alternatives departs from ECMA-262 22.2.2.7 CompileAtom and 22.2.6.11 RegExp.prototype [ Symbol.replace ] / AdvanceStringIndex; native and JavaScript must agree`},
}

func TestRegexCoverageRefusals(t *testing.T) {
	t.Parallel()
	for _, fixture := range regexCoverageRefusals {
		t.Run(fixture.file, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(repository, "internal/oracle/testdata/regex_coverage_refused", fixture.file+".a")
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), fixture.message) {
				t.Fatalf("want %q, got %v", fixture.message, err)
			}
		})
	}
}
