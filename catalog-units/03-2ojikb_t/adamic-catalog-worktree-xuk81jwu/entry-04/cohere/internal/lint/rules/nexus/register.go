package nexus

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
		rule.Registration{Rule: BoundaryNoInternalImport},
		rule.Registration{Rule: BoundaryNoNexusOutsideImport},
		rule.Registration{
			Rule:            BoundaryNoProjectImport,
			Decode:          rule.DecodeOptionsInto[BoundaryNoProjectImportOptions](),
			RequiresOptions: true,
		},
		rule.Registration{
			Rule:   ConsistencyNoAbbreviatedIdentifier,
			Decode: rule.DecodeOptionsInto[ConsistencyNoAbbreviatedIdentifierOptions](),
		},
		rule.Registration{Rule: ConsistencyNoAmbiguousIdentifier},
		rule.Registration{
			Rule:   ConsistencyNoBooleanOutcome,
			Decode: rule.DecodeOptionsInto[ConsistencyNoBooleanOutcomeOptions](),
		},
		rule.Registration{Rule: ConsistencyNoEnum},
		rule.Registration{
			Rule:   ConsistencyNoLongLineComment,
			Decode: rule.DecodeOptionsInto[ConsistencyNoLongLineCommentOptions](),
		},
		rule.Registration{Rule: ConsistencyNoMultilineArrowFunction},
		rule.Registration{
			Rule:   ConsistencyNoScreamingSnakeCase,
			Decode: rule.DecodeOptionsInto[ConsistencyNoScreamingSnakeCaseOptions](),
		},
		rule.Registration{
			Rule:   ConsistencyNoShouting,
			Decode: rule.DecodeOptionsInto[ConsistencyNoShoutingOptions](),
		},
		rule.Registration{Rule: ConsistencyNoSingleLineJsDoc},
		rule.Registration{
			Rule:   ConsistencyNoStutteringName,
			Decode: rule.DecodeOptionsInto[ConsistencyNoStutteringNameOptions](),
		},
		rule.Registration{Rule: ConsistencyNoUtilsFolder},
		rule.Registration{
			Rule:   ConsistencyRequireConstantCasing,
			Decode: rule.DecodeOptionsInto[ConsistencyRequireConstantCasingOptions](),
		},
		rule.Registration{Rule: ConsistencyRequireTypeSuffix},
		rule.Registration{Rule: ImportNoForbiddenSource},
		rule.Registration{
			Rule:   ImportRequireModuleAlias,
			Decode: rule.DecodeOptionsInto[ImportRequireModuleAliasOptions](),
		},
		rule.Registration{Rule: ImportRequireNodeNamespace},
		rule.Registration{
			Rule:            ImportRequirePathAlias,
			DecodeAt:        decodeImportRequirePathAliasOptions,
			RequiresOptions: true,
		},
		rule.Registration{Rule: LocalizationNoUntranslatedValue},
	)
}
