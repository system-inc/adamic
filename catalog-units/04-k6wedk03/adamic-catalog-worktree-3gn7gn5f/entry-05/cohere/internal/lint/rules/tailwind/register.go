package tailwind

import (
	"github.com/system-inc/cohere/internal/lint/rule"
)

// init registers this package's rules with the catalog.
//
// Registration lives beside the rules rather than in a shared list another package maintains. That
// shared list was one file every port had to edit, so two ports landing at once conflicted on a
// file neither was really changing, and the conflict scaled with how many rules were being written
// at the same time. A rule is now two new files and no shared edit.
//
// The cost is that a package whose rules are never imported registers nothing, so the catalog would
// be silently short. `registry` imports every rule package for exactly that reason, and the parity
// guard reads the result against the inventory, so a package dropped from those imports shows up as
// rules that vanished rather than as a quieter run.
func init() {
	rule.Register(
		tailwindRegistration[EnforceConsistentClassOrderOptions](EnforceConsistentClassOrder),
		tailwindRegistration[EnforceConsistentVariantOrderOptions](EnforceConsistentVariantOrder),
		tailwindRegistration[EnforceConsistentImportantPositionOptions](EnforceConsistentImportantPosition),
		tailwindRegistration[EnforceConsistentVariableSyntaxOptions](EnforceConsistentVariableSyntax),
		tailwindRegistration[EnforceShorthandClassesOptions](EnforceShorthandClasses),
		tailwindRegistration[EnforceCanonicalClassesOptions](EnforceCanonicalClasses),
		tailwindRegistration[NoConcatenatedClassesOptions](NoConcatenatedClasses),
		tailwindRegistration[NoConflictingClassesOptions](NoConflictingClasses),
		tailwindRegistration[NoDeprecatedClassesOptions](NoDeprecatedClasses),
		tailwindRegistration[NoDuplicateClassesOptions](NoDuplicateClasses),
		tailwindRegistration[NoUnknownClassesOptions](NoUnknownClasses),
		tailwindRegistration[NoUnnecessaryWhitespaceOptions](NoUnnecessaryWhitespace),
		rule.Registration{Rule: NoPhysicalDirection},
	)
	rule.RegisterSettings(rule.SettingsRegistration{
		Namespaces: tailwindSettingsNamespaces,
		Split:      splitTailwindSettings,
	})
}
