package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// An Adamic `.a` translation file is checked as the `.ts` it is, against its English sibling found by
// its own extension first and then the other (#kwt1htp). A set moving to Adamic file by file holds
// both kinds side by side, so each pairing finds its English. Both siblings present is a guess, so it
// is reported and nothing is compared; neither present is reported as missing, naming en.a first
// (#techtr1).
func TestAnAdamicTranslationFileFindsItsEnglishSibling(t *testing.T) {
	t.Parallel()
	english := "export default { Greeting: 'Hello there' };\n"
	copied := "export default { Greeting: 'Hello there' };\n"

	control := runOnTranslations(t, map[string]string{
		"_translations/en.ts": english,
		"_translations/fr.ts": copied,
	}, "_translations/fr.ts")

	for _, testCase := range []struct {
		name  string
		files map[string]string
	}{
		{"en.a beside fr.a", map[string]string{"_translations/en.a": english, "_translations/fr.a": copied}},
		{"en.ts beside fr.a", map[string]string{"_translations/en.ts": english, "_translations/fr.a": copied}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectSameFindings(t, control, runOnTranslations(t, testCase.files, "_translations/fr.a"))
		})
	}

	t.Run("en.a beside fr.ts", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectSameFindings(t, control, runOnTranslations(t, map[string]string{
			"_translations/en.a":  english,
			"_translations/fr.ts": copied,
		}, "_translations/fr.ts"))
	})

	t.Run("both en.ts and en.a", func(t *testing.T) {
		t.Parallel()
		result := runOnTranslations(t, map[string]string{
			"_translations/en.ts": english,
			"_translations/en.a":  english,
			"_translations/fr.a":  copied,
		}, "_translations/fr.a")
		rule_testing.ExpectFindings(t, result, "ambiguousEnglishSibling")
	})

	t.Run("neither en.ts nor en.a", func(t *testing.T) {
		t.Parallel()
		result := runOnTranslations(t, map[string]string{
			"_translations/de.ts": english,
			"_translations/fr.a":  copied,
		}, "_translations/fr.a")
		rule_testing.ExpectFindings(t, result, "missingEnglishSibling")
		expectMissingSiblingNames(t, result, "en.a", "en.ts")
	})

	// en.a is English, not a locale to check, as en.ts is.
	t.Run("en.a itself", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runOnTranslations(t, map[string]string{
			"_translations/en.ts": english,
			"_translations/en.a":  copied,
		}, "_translations/en.a"))
	})
}
