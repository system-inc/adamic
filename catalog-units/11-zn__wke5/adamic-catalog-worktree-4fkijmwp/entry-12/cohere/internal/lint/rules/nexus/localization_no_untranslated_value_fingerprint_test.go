package nexus

import (
	"crypto/sha256"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// localizationNoUntranslatedValueFingerprintOf is the rule's program fingerprint over a fixture, read through
// the Program the walk hands the rule, viewed under its own reads. Each fixture is built in a directory of its
// own, so two that agree prove the tables are named relative to the root rather than where they were built.
func localizationNoUntranslatedValueFingerprintOf(t *testing.T, files map[string]string) [sha256.Size]byte {
	t.Helper()
	graph, _ := buildTranslations(t, files)
	return LocalizationNoUntranslatedValue.ProgramFingerprint(rule.ViewProgram(graph.Program, nil, LocalizationNoUntranslatedValue), nil)
}

// The fingerprint is every English table a locale file could read (#31ffaaa), proven both ways: a table's
// text, its presence or its name changing moves it, en.a beside en.ts and a set gaining its first table
// included, since both change what a locale file reports; and an edit no locale file's verdict reads leaves
// it, so the findings cache replays across it.
func TestLocalizationNoUntranslatedValueFingerprintsExactlyTheEnglishTables(t *testing.T) {
	t.Parallel()
	const english = "source/translations/en.ts"
	base := map[string]string{
		english:                           englishTranslations,
		"source/translations/es.ts":       "export default {\n    Greeting: 'Hola',\n};\n",
		"source/admin/translations/de.ts": "export default {\n    Greeting: 'Hallo',\n};\n",
		"source/en.ts":                    "export default { Greeting: 'Hello there' };\n",
		"source/lib/util.ts":              "export const util = 1;\n",
	}
	edited := func(changes map[string]string, removed ...string) map[string]string {
		files := map[string]string{}
		for name, contents := range base {
			files[name] = contents
		}
		for _, name := range removed {
			delete(files, name)
		}
		for name, contents := range changes {
			files[name] = contents
		}
		return files
	}
	original := localizationNoUntranslatedValueFingerprintOf(t, edited(nil))

	moves := map[string]map[string]string{
		"an English value edited":                edited(map[string]string{english: "export default {\n    Greeting: 'Hi there',\n};\n"}),
		"a comment in en.ts":                     edited(map[string]string{english: "// The English set.\n" + englishTranslations}),
		"en.ts deleted":                          edited(nil, english),
		"en.a added beside en.ts":                edited(map[string]string{"source/translations/en.a": englishTranslations}),
		"en.ts renamed to en.a":                  edited(map[string]string{"source/translations/en.a": englishTranslations}, english),
		"a set gaining its first en.ts":          edited(map[string]string{"source/admin/translations/en.ts": englishTranslations}),
		"an en.ts in an _translations set":       edited(map[string]string{"source/billing/_translations/en.ts": englishTranslations}),
		"en.ts moved to another translation set": edited(map[string]string{"source/admin/translations/en.ts": englishTranslations}, english),
	}
	for name, files := range moves {
		if localizationNoUntranslatedValueFingerprintOf(t, files) == original {
			t.Errorf("%s left the fingerprint unchanged, so a stale finding would replay", name)
		}
	}

	holds := map[string]map[string]string{
		"nothing changed, built elsewhere":     edited(nil),
		"a locale file edited":                 edited(map[string]string{"source/translations/es.ts": "export default {\n    Greeting: 'Hello there',\n};\n"}),
		"a locale file added to a set":         edited(map[string]string{"source/translations/fr.ts": "export default {\n    Greeting: 'Bonjour',\n};\n"}),
		"an unrelated file edited":             edited(map[string]string{"source/lib/util.ts": "export const util = 2;\n"}),
		"an en.ts outside any translation set": edited(map[string]string{"source/en.ts": "export default { Greeting: 'Hi' };\n"}),
		"an en.ts in a translations-archive":   edited(map[string]string{"source/translations-archive/en.ts": englishTranslations}),
	}
	for name, files := range holds {
		if localizationNoUntranslatedValueFingerprintOf(t, files) != original {
			t.Errorf("%s moved the fingerprint, so files whose verdict cannot change would not replay", name)
		}
	}
}
