package core

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's no-misleading-character-class, run through the installed rule under @typescript-eslint/parser
 * over its whole test corpus and over the edges this port's own reading raised (#jjfa7qb), with every
 * finding's id, text and byte offset written here as ESLint answered.
 *
 * The corpus rows are the registry's corpus file, verbatim, less the rows ESLint runs under configured
 * globals, the rows that set allowEscape, which TestNoMisleadingCharacterClassAllowEscape runs both ways
 * round, and the rows the registry's known-gaps list holds for this rule, which that list keeps exactly.
 * The edge rows were written for this port: a RegExp reached through an alias or globalThis and a local
 * one that is not it, a pattern held in a name or built by concatenation, a regex bound to a constant,
 * regex literals handed over with and without flags, and flags that are constant, written, or
 * unknowable.
 */
func TestNoMisleadingCharacterClassAgreesWithESLint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		origin string
		code   string
		want   []string
	}{
		{"corpus", "var r = /[👍]/u", []string{}},
		{"corpus", "var r = /[\\uD83D\\uDC4D]/u", []string{}},
		{"corpus", "var r = /[\\u{1F44D}]/u", []string{}},
		{"corpus", "var r = /❇️/", []string{}},
		{"corpus", "var r = /Á/", []string{}},
		{"corpus", "var r = /[❇]/", []string{}},
		{"corpus", "var r = /👶🏻/", []string{}},
		{"corpus", "var r = /[👶]/u", []string{}},
		{"corpus", "var r = /🇯🇵/", []string{}},
		{"corpus", "var r = /[JP]/", []string{}},
		{"corpus", "var r = /👨‍👩‍👦/", []string{}},
		{"corpus", "new RegExp()", []string{}},
		{"corpus", "var r = RegExp(/[👍]/u)", []string{}},
		{"corpus", "const regex = /[👍]/u; new RegExp(regex);", []string{}},
		{"corpus", "var r = /[\\uD83D]/", []string{}},
		{"corpus", "var r = /[\\uDC4D]/", []string{}},
		{"corpus", "var r = /[\\uD83D]/u", []string{}},
		{"corpus", "var r = /[\\uDC4D]/u", []string{}},
		{"corpus", "var r = /[\\u0301]/", []string{}},
		{"corpus", "var r = /[\\uFE0F]/", []string{}},
		{"corpus", "var r = /[\\u0301]/u", []string{}},
		{"corpus", "var r = /[\\uFE0F]/u", []string{}},
		{"corpus", "var r = /[\\u{1F3FB}]/u", []string{}},
		{"corpus", "var r = /[🏻]/u", []string{}},
		{"corpus", "var r = /[🇯]/u", []string{}},
		{"corpus", "var r = /[🇵]/u", []string{}},
		{"corpus", "var r = /[\\u200D]/", []string{}},
		{"corpus", "var r = /[\\u200D]/u", []string{}},
		{"corpus", "new RegExp('[Á] [ ');", []string{}},
		{"corpus", "var r = new RegExp('[Á] [ ');", []string{}},
		{"corpus", "var r = RegExp('{ [Á]', 'u');", []string{}},
		{"corpus", "var r = new globalThis.RegExp('[Á] [ ');", []string{}},
		{"corpus", "var r = globalThis.RegExp('{ [Á]', 'u');", []string{}},
		{"corpus", "var r = RegExp(`${x}[👍]`)", []string{}},
		{"corpus", "var r = new RegExp('[🇯🇵]', `${foo}`)", []string{}},
		{"corpus", "var r = new RegExp(\"[👍]\", flags)", []string{}},
		{"corpus", "const args = ['[👍]', 'i']; new RegExp(...args);", []string{}},
		{"corpus", "var r = /[👍]/v", []string{}},
		{"corpus", "var r = /^[\\q{👶🏻}]$/v", []string{}},
		{"corpus", "var r = /[🇯\\q{abc}🇵]/v", []string{}},
		{"corpus", "var r = /[🇯[A]🇵]/v", []string{}},
		{"corpus", "var r = /[🇯[A--B]🇵]/v", []string{}},
		{"corpus", "var r = /[👍]/", []string{"surrogatePairWithoutUFlag 👍@10"}},
		{"corpus", "var r = /[\\uD83D\\uDC4D]/", []string{"surrogatePairWithoutUFlag \\uD83D\\uDC4D@10"}},
		{"corpus", "var r = /[👍]/", []string{"surrogatePairWithoutUFlag 👍@10"}},
		{"corpus", "var r = /before[\\uD83D\\uDC4D]after/", []string{"surrogatePairWithoutUFlag \\uD83D\\uDC4D@16"}},
		{"corpus", "var r = /[before\\uD83D\\uDC4Dafter]/", []string{"surrogatePairWithoutUFlag \\uD83D\\uDC4D@16"}},
		{"corpus", "var r = /\\uDC4D[\\uD83D\\uDC4D]/", []string{"surrogatePairWithoutUFlag \\uD83D\\uDC4D@16"}},
		{"corpus", "var r = /[👍]/", []string{"surrogatePairWithoutUFlag 👍@10"}},
		{"corpus", "var r = /[👍]\\a/", []string{"surrogatePairWithoutUFlag 👍@10"}},
		{"corpus", "var r = /\\a[👍]\\a/", []string{"surrogatePairWithoutUFlag 👍@12"}},
		{"corpus", "var r = /(?<=[👍])/", []string{"surrogatePairWithoutUFlag 👍@14"}},
		{"corpus", "var r = /(?<=[👍])/", []string{"surrogatePairWithoutUFlag 👍@14"}},
		{"corpus", "var r = /[Á]/", []string{"combiningClass Á@10"}},
		{"corpus", "var r = /[Á]/u", []string{"combiningClass Á@10"}},
		{"corpus", "var r = /[\\u0041\\u0301]/", []string{"combiningClass \\u0041\\u0301@10"}},
		{"corpus", "var r = /[\\u0041\\u0301]/u", []string{"combiningClass \\u0041\\u0301@10"}},
		{"corpus", "var r = /[\\u{41}\\u{301}]/u", []string{"combiningClass \\u{41}\\u{301}@10"}},
		{"corpus", "var r = /[❇️]/", []string{"combiningClass ❇️@10"}},
		{"corpus", "var r = /[❇️]/u", []string{"combiningClass ❇️@10"}},
		{"corpus", "var r = /[\\u2747\\uFE0F]/", []string{"combiningClass \\u2747\\uFE0F@10"}},
		{"corpus", "var r = /[\\u2747\\uFE0F]/u", []string{"combiningClass \\u2747\\uFE0F@10"}},
		{"corpus", "var r = /[\\u{2747}\\u{FE0F}]/u", []string{"combiningClass \\u{2747}\\u{FE0F}@10"}},
		{"corpus", "var r = /[👶🏻]/", []string{"surrogatePairWithoutUFlag 👶@10", "surrogatePairWithoutUFlag 🏻@14"}},
		{"corpus", "var r = /[👶🏻]/u", []string{"emojiModifier 👶🏻@10"}},
		{"corpus", "var r = /[a\\uD83C\\uDFFB]/u", []string{"emojiModifier a\\uD83C\\uDFFB@10"}},
		{"corpus", "var r = /[\\uD83D\\uDC76\\uD83C\\uDFFB]/u", []string{"emojiModifier \\uD83D\\uDC76\\uD83C\\uDFFB@10"}},
		{"corpus", "var r = /[\\u{1F476}\\u{1F3FB}]/u", []string{"emojiModifier \\u{1F476}\\u{1F3FB}@10"}},
		{"corpus", "var r = /[🇯🇵]/", []string{"surrogatePairWithoutUFlag 🇯@10", "surrogatePairWithoutUFlag 🇵@14"}},
		{"corpus", "var r = /[🇯🇵]/i", []string{"surrogatePairWithoutUFlag 🇯@10", "surrogatePairWithoutUFlag 🇵@14"}},
		{"corpus", "var r = /[🇯🇵]/u", []string{"regionalIndicatorSymbol 🇯🇵@10"}},
		{"corpus", "var r = /[\\uD83C\\uDDEF\\uD83C\\uDDF5]/u", []string{"regionalIndicatorSymbol \\uD83C\\uDDEF\\uD83C\\uDDF5@10"}},
		{"corpus", "var r = /[\\u{1F1EF}\\u{1F1F5}]/u", []string{"regionalIndicatorSymbol \\u{1F1EF}\\u{1F1F5}@10"}},
		{"corpus", "var r = /[👨‍👩‍👦]/u", []string{"zwj 👨‍👩‍👦@10"}},
		{"corpus", "var r = /[👩‍👦]/u", []string{"zwj 👩‍👦@10"}},
		{"corpus", "var r = /[👩‍👦][👩‍👦]/u", []string{"zwj 👩‍👦@10", "zwj 👩‍👦@23"}},
		{"corpus", "var r = /[👨‍👩‍👦]foo[👨‍👩‍👦]/u", []string{"zwj 👨‍👩‍👦@10", "zwj 👨‍👩‍👦@33"}},
		{"corpus", "var r = /[👨‍👩‍👦👩‍👦]/u", []string{"zwj 👨‍👩‍👦@10", "zwj 👩‍👦@28"}},
		{"corpus", "var r = /[\\uD83D\\uDC68\\u200D\\uD83D\\uDC69\\u200D\\uD83D\\uDC66]/u", []string{"zwj \\uD83D\\uDC68\\u200D\\uD83D\\uDC69\\u200D\\uD83D\\uDC66@10"}},
		{"corpus", "var r = /[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]/u", []string{"zwj \\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}@10"}},
		{"corpus", "var r = /[\\uD83D\\uDC68\\u200D\\uD83D\\uDC69]/u", []string{"zwj \\uD83D\\uDC68\\u200D\\uD83D\\uDC69@10"}},
		{"corpus", "var r = /[\\u{1F468}\\u{200D}\\u{1F469}]/u", []string{"zwj \\u{1F468}\\u{200D}\\u{1F469}@10"}},
		{"corpus", "var r = /[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]foo[\\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}]/u", []string{"zwj \\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}@10", "zwj \\u{1F468}\\u{200D}\\u{1F469}\\u{200D}\\u{1F466}@58"}},
		{"corpus", "var r = RegExp(\"[👍]\", \"\")", []string{"surrogatePairWithoutUFlag 👍@17"}},
		{"corpus", "var r = new RegExp(\"[👍]\", \"\")", []string{"surrogatePairWithoutUFlag 👍@21"}},
		{"corpus", "var r = new RegExp('[👍]', ``)", []string{"surrogatePairWithoutUFlag 👍@21"}},
		{"corpus", "var r = new RegExp(`\n                [👍]`)", []string{"surrogatePairWithoutUFlag 👍@38"}},
		{"corpus", "var r = new RegExp(`\n                [❇️]`)", []string{"combiningClass ❇️@38"}},
		{"corpus", "const flags = \"\"; var r = new RegExp(\"[👍]\", flags)", []string{"surrogatePairWithoutUFlag 👍@39"}},
		{"corpus", "var r = RegExp(\"[\\\\uD83D\\\\uDC4D]\", \"\")", []string{"surrogatePairWithoutUFlag \\\\uD83D\\\\uDC4D@17"}},
		{"corpus", "var r = RegExp(\"before[\\\\uD83D\\\\uDC4D]after\", \"\")", []string{"surrogatePairWithoutUFlag \\\\uD83D\\\\uDC4D@23"}},
		{"corpus", "var r = RegExp(\"[before\\\\uD83D\\\\uDC4Dafter]\", \"\")", []string{"surrogatePairWithoutUFlag \\\\uD83D\\\\uDC4D@23"}},
		{"corpus", "var r = RegExp(\"\\t\\t\\t👍[👍]\")", []string{"surrogatePairWithoutUFlag 👍@27"}},
		{"corpus", "var r = new RegExp(\"\\u1234[\\\\uD83D\\\\uDC4D]\")", []string{"surrogatePairWithoutUFlag \\\\uD83D\\\\uDC4D@27"}},
		{"corpus", "var r = new RegExp(\"\\\\u1234\\\\u5678👎[👍]\")", []string{"surrogatePairWithoutUFlag 👍@39"}},
		{"corpus", "var r = new RegExp(\"\\\\u1234\\\\u5678👍[👍]\")", []string{"surrogatePairWithoutUFlag 👍@39"}},
		{"corpus", "var r = new RegExp(\"[👍]\", \"\")", []string{"surrogatePairWithoutUFlag 👍@21"}},
		{"corpus", "var r = new RegExp(\"[👍]\", \"\")", []string{"surrogatePairWithoutUFlag 👍@21"}},
		{"corpus", "var r = new RegExp(\"[👍]\\\\a\", \"\")", []string{"surrogatePairWithoutUFlag 👍@21"}},
		{"corpus", "var r = new RegExp(\"/(?<=[👍])\", \"\")", []string{"surrogatePairWithoutUFlag 👍@26"}},
		{"corpus", "var r = new RegExp(\"/(?<=[👍])\", \"\")", []string{"surrogatePairWithoutUFlag 👍@26"}},
		{"corpus", "var r = new RegExp(\"[Á]\", \"\")", []string{"combiningClass Á@21"}},
		{"corpus", "var r = new RegExp(\"[Á]\", \"u\")", []string{"combiningClass Á@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u0041\\\\u0301]\", \"\")", []string{"combiningClass \\\\u0041\\\\u0301@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u0041\\\\u0301]\", \"u\")", []string{"combiningClass \\\\u0041\\\\u0301@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u{41}\\\\u{301}]\", \"u\")", []string{"combiningClass \\\\u{41}\\\\u{301}@21"}},
		{"corpus", "var r = new RegExp(\"[❇️]\", \"\")", []string{"combiningClass ❇️@21"}},
		{"corpus", "var r = new RegExp(\"[❇️]\", \"u\")", []string{"combiningClass ❇️@21"}},
		{"corpus", "new RegExp(\"[ \\\\ufe0f]\", \"\")", []string{"combiningClass  \\\\ufe0f@13"}},
		{"corpus", "new RegExp(\"[ \\\\ufe0f]\", \"u\")", []string{"combiningClass  \\\\ufe0f@13"}},
		{"corpus", "new RegExp(\"[ \\\\ufe0f][ \\\\ufe0f]\")", []string{"combiningClass  \\\\ufe0f@13", "combiningClass  \\\\ufe0f@23"}},
		{"corpus", "var r = new RegExp(\"[\\\\u2747\\\\uFE0F]\", \"\")", []string{"combiningClass \\\\u2747\\\\uFE0F@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u2747\\\\uFE0F]\", \"u\")", []string{"combiningClass \\\\u2747\\\\uFE0F@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u{2747}\\\\u{FE0F}]\", \"u\")", []string{"combiningClass \\\\u{2747}\\\\u{FE0F}@21"}},
		{"corpus", "var r = new RegExp(\"[👶🏻]\", \"\")", []string{"surrogatePairWithoutUFlag 👶@21", "surrogatePairWithoutUFlag 🏻@25"}},
		{"corpus", "var r = new RegExp(\"[👶🏻]\", \"u\")", []string{"emojiModifier 👶🏻@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\uD83D\\\\uDC76\\\\uD83C\\\\uDFFB]\", \"u\")", []string{"emojiModifier \\\\uD83D\\\\uDC76\\\\uD83C\\\\uDFFB@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u{1F476}\\\\u{1F3FB}]\", \"u\")", []string{"emojiModifier \\\\u{1F476}\\\\u{1F3FB}@21"}},
		{"corpus", "var r = RegExp(`\t\t\t👍[👍]`)", []string{"surrogatePairWithoutUFlag 👍@24"}},
		{"corpus", "var r = RegExp(`\\t\\t\\t👍[👍]`)", []string{"surrogatePairWithoutUFlag 👍@27"}},
		{"corpus", "var r = new RegExp(\"[🇯🇵]\", \"\")", []string{"surrogatePairWithoutUFlag 🇯@21", "surrogatePairWithoutUFlag 🇵@25"}},
		{"corpus", "var r = new RegExp(\"[🇯🇵]\", \"i\")", []string{"surrogatePairWithoutUFlag 🇯@21", "surrogatePairWithoutUFlag 🇵@25"}},
		{"corpus", "var r = new RegExp('[🇯🇵]', `i`)", []string{"surrogatePairWithoutUFlag 🇯@21", "surrogatePairWithoutUFlag 🇵@25"}},
		{"corpus", "var r = new RegExp(\"[🇯🇵]\")", []string{"surrogatePairWithoutUFlag 🇯@21", "surrogatePairWithoutUFlag 🇵@25"}},
		{"corpus", "var r = new RegExp(\"[🇯🇵]\",)", []string{"surrogatePairWithoutUFlag 🇯@21", "surrogatePairWithoutUFlag 🇵@25"}},
		{"corpus", "var r = new RegExp((\"[🇯🇵]\"))", []string{"surrogatePairWithoutUFlag 🇯@22", "surrogatePairWithoutUFlag 🇵@26"}},
		{"corpus", "var r = new RegExp(((\"[🇯🇵]\")))", []string{"surrogatePairWithoutUFlag 🇯@23", "surrogatePairWithoutUFlag 🇵@27"}},
		{"corpus", "var r = new RegExp((\"[🇯🇵]\"),)", []string{"surrogatePairWithoutUFlag 🇯@22", "surrogatePairWithoutUFlag 🇵@26"}},
		{"corpus", "var r = new RegExp(\"[🇯🇵]\", \"u\")", []string{"regionalIndicatorSymbol 🇯🇵@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\uD83C\\\\uDDEF\\\\uD83C\\\\uDDF5]\", \"u\")", []string{"regionalIndicatorSymbol \\\\uD83C\\\\uDDEF\\\\uD83C\\\\uDDF5@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u{1F1EF}\\\\u{1F1F5}]\", \"u\")", []string{"regionalIndicatorSymbol \\\\u{1F1EF}\\\\u{1F1F5}@21"}},
		{"corpus", "var r = new RegExp(\"[👨‍👩‍👦]\", \"u\")", []string{"zwj 👨‍👩‍👦@21"}},
		{"corpus", "var r = new RegExp(\"[👩‍👦]\", \"u\")", []string{"zwj 👩‍👦@21"}},
		{"corpus", "var r = new RegExp(\"[👩‍👦][👩‍👦]\", \"u\")", []string{"zwj 👩‍👦@21", "zwj 👩‍👦@34"}},
		{"corpus", "var r = new RegExp(\"[👨‍👩‍👦]foo[👨‍👩‍👦]\", \"u\")", []string{"zwj 👨‍👩‍👦@21", "zwj 👨‍👩‍👦@44"}},
		{"corpus", "var r = new RegExp(\"[👨‍👩‍👦👩‍👦]\", \"u\")", []string{"zwj 👨‍👩‍👦@21", "zwj 👩‍👦@39"}},
		{"corpus", "var r = new RegExp(\"[\\\\uD83D\\\\uDC68\\\\u200D\\\\uD83D\\\\uDC69\\\\u200D\\\\uD83D\\\\uDC66]\", \"u\")", []string{"zwj \\\\uD83D\\\\uDC68\\\\u200D\\\\uD83D\\\\uDC69\\\\u200D\\\\uD83D\\\\uDC66@21"}},
		{"corpus", "var r = new RegExp(\"[\\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}]\", \"u\")", []string{"zwj \\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}@21"}},
		{"corpus", "var r = new globalThis.RegExp(\"[❇️]\", \"\")", []string{"combiningClass ❇️@32"}},
		{"corpus", "var r = new globalThis.RegExp(\"[👶🏻]\", \"u\")", []string{"emojiModifier 👶🏻@32"}},
		{"corpus", "var r = new globalThis.RegExp(\"[🇯🇵]\", \"\")", []string{"surrogatePairWithoutUFlag 🇯@32", "surrogatePairWithoutUFlag 🇵@36"}},
		{"corpus", "var r = new globalThis.RegExp(\"[\\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}]\", \"u\")", []string{"zwj \\\\u{1F468}\\\\u{200D}\\\\u{1F469}\\\\u{200D}\\\\u{1F466}@32"}},
		{"corpus", "/[\\ud83d\\u{dc4d}]/u", []string{"surrogatePair \\ud83d\\u{dc4d}@2"}},
		{"corpus", "/[\\u{d83d}\\udc4d]/u", []string{"surrogatePair \\u{d83d}\\udc4d@2"}},
		{"corpus", "/[\\u{d83d}\\u{dc4d}]/u", []string{"surrogatePair \\u{d83d}\\u{dc4d}@2"}},
		{"corpus", "/[\\uD83D\\u{DC4d}]/u", []string{"surrogatePair \\uD83D\\u{DC4d}@2"}},
		{"corpus", "new RegExp(`${\"[👍🇯🇵]\"}[😊]`);", []string{"surrogatePairWithoutUFlag `${\"[👍🇯🇵]\"}[😊]`@11"}},
		{"corpus", "const pattern = \"[👍]\"; new RegExp(pattern);", []string{"surrogatePairWithoutUFlag pattern@37"}},
		{"corpus", "RegExp(/[a👍z]/u, '');", []string{"surrogatePairWithoutUFlag 👍@10"}},
		{"corpus", "RegExp(/[👍]/)", []string{"surrogatePairWithoutUFlag 👍@9"}},
		{"corpus", "RegExp(/[👍]/, 'i');", []string{"surrogatePairWithoutUFlag 👍@9"}},
		{"corpus", "\n\n            // \"[\" and \"]\" escaped as \"\\x5B\" and \"\\u005D\"\n            new RegExp(\"\\x5B \\\\ufe0f\\u005D\")\n\n            ", []string{"combiningClass  \\\\ufe0f@88"}},
		{"corpus", "\n\n            // backslash escaped as \"\\u{5c}\"\n            new RegExp(\"[ \\u{5c}ufe0f]\")\n\n            ", []string{"combiningClass  \\u{5c}ufe0f@72"}},
		{"corpus", "\n\n            // \"e\" escaped as \"\\e\"\n            new RegExp(\"[ \\\\uf\\e0f]\")\n\n            ", []string{"combiningClass  \\\\uf\\e0f@62"}},
		{"corpus", "\n\n            // just a backslash escaped as \"\\\\\"\n            new RegExp(`[.\\\\u200D.]`)\n\n            ", []string{"zwj .\\\\u200D.@75"}},
		{"corpus", "\n\n            // \"u\" escaped as \"\\x75\"\n            new RegExp(`[.\\\\\\x75200D.]`)\n\n            ", []string{"zwj .\\\\\\x75200D.@64"}},
		{"corpus", "\n\n            // unescaped <CR> <LF> counts as a single character\n            new RegExp(`[\n\\\\u200D.]`)\n\n            ", []string{"zwj \n\\\\u200D.@91"}},
		{"corpus", "var r = /[[👶🏻]]/v", []string{"emojiModifier 👶🏻@11"}},
		{"corpus", "new RegExp(/^[👍]$/v, '')", []string{"surrogatePairWithoutUFlag 👍@14"}},
		{"corpus", "var r = /[\\uD83D\\uDC4D-\\uffff]/", []string{"surrogatePairWithoutUFlag \\uD83D\\uDC4D@10"}},
		{"corpus", "var r = /[👨‍👩‍👦]/", []string{"surrogatePairWithoutUFlag 👨@10", "zwj 👨‍@10", "surrogatePairWithoutUFlag 👩@17", "zwj 👩‍@17", "surrogatePairWithoutUFlag 👦@24"}},
		{"corpus", "var r = new RegExp(\"[👨‍👩‍👦]\", \"\")", []string{"surrogatePairWithoutUFlag 👨@21", "zwj 👨‍@21", "surrogatePairWithoutUFlag 👩@28", "zwj 👩‍@28", "surrogatePairWithoutUFlag 👦@35"}},
		{"edge", "const R = RegExp; new R(\"[👍]\");", []string{"surrogatePairWithoutUFlag 👍@26"}},
		{"edge", "globalThis.RegExp(\"[👍]\");", []string{"surrogatePairWithoutUFlag 👍@20"}},
		{"edge", "let RegExp; new RegExp(\"[👍]\");", []string{}},
		{"edge", "function f(RegExp) { return new RegExp(\"[👍]\"); }", []string{}},
		{"edge", "const p = \"[👍]\"; new RegExp(p);", []string{"surrogatePairWithoutUFlag p@31"}},
		{"edge", "new RegExp(\"[\" + \"👍]\");", []string{"surrogatePairWithoutUFlag \"[\" + \"👍]\"@11"}},
		{"edge", "const p = \"[👍][👍]\"; new RegExp(p);", []string{"surrogatePairWithoutUFlag p@37"}},
		{"edge", "const r = /[👍]/; new RegExp(r, \"\");", []string{"surrogatePairWithoutUFlag 👍@12"}},
		{"edge", "new RegExp(/[👍]/u, \"\");", []string{"surrogatePairWithoutUFlag 👍@13"}},
		{"edge", "new RegExp(/[👍]/u);", []string{}},
		{"edge", "new RegExp(/[👍]/, flags);", []string{}},
		{"edge", "const f = \"u\"; new RegExp(\"[👍]\", f);", []string{}},
		{"edge", "let f = \"u\"; f = \"\"; new RegExp(\"[👍]\", f);", []string{}},
		{"edge", "new RegExp(\"[👍]\", `${f}`);", []string{}},
		{"edge", "new RegExp(`[👍]`);", []string{"surrogatePairWithoutUFlag 👍@13"}},
		{"edge", "new RegExp((\"[👍]\"));", []string{"surrogatePairWithoutUFlag 👍@14"}},
		{"edge", "const p = \"[👍]\"; const q = p; new RegExp(q, \"\");", []string{"surrogatePairWithoutUFlag q@44"}},
		{"edge", "new RegExp(pattern);", []string{}},
	}

	for index, testCase := range cases {
		result := rule_testing.RunTypedVerbatimWithOptions(t, NoMisleadingCharacterClass, "input.ts", testCase.code, nil)
		source := result.SourceFile.Text()
		got := []string{}
		for _, diagnostic := range result.Diagnostics {
			got = append(got, fmt.Sprintf("%s %s@%d", diagnostic.Message.Id,
				source[diagnostic.Range.Pos():diagnostic.Range.End()], diagnostic.Range.Pos()))
		}
		want := append([]string{}, testCase.want...)
		sort.Strings(got)
		sort.Strings(want)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("case %d (%s): %q\n got: %q\nwant: %q", index, testCase.origin, testCase.code, got, want)
		}
	}
}
