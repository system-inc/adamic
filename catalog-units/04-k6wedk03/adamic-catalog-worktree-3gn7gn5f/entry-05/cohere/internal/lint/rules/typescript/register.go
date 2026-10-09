package typescript

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
		rule.Registration{Rule: AwaitThenable},
		rule.Registration{
			Rule:   BanTsComment,
			Decode: DecodeBanTsCommentOptions,
		},
		rule.Registration{
			Rule:   ConsistentTypeImports,
			Decode: DecodeConsistentTypeImportsOptions,
		},
		rule.Registration{Rule: NoArrayDelete},
		rule.Registration{Rule: NoDuplicateEnumValues},
		rule.Registration{
			Rule:   NoExplicitAny,
			Decode: rule.DecodeOptionsInto[NoExplicitAnyOptions](),
		},
		rule.Registration{Rule: NoExtraNonNullAssertion},
		rule.Registration{Rule: NoImpliedEval},
		rule.Registration{
			Rule:   NoFloatingPromises,
			Decode: DecodeNoFloatingPromisesOptions,
		},
		rule.Registration{Rule: NoForInArray},
		rule.Registration{Rule: NoMisusedNew},
		rule.Registration{
			Rule:   NoMisusedPromises,
			Decode: rule.DecodeOptionsInto[NoMisusedPromisesOptions](),
		},
		rule.Registration{Rule: NoNonNullAssertedOptionalChain},
		rule.Registration{Rule: NoNonNullAssertion},
		rule.Registration{
			Rule:   NoRequireImports,
			Decode: DecodeNoRequireImportsOptions,
		},
		rule.Registration{
			Rule:   NoThisAlias,
			Decode: DecodeNoThisAliasOptions,
		},
		rule.Registration{Rule: NoUnnecessaryParameterPropertyAssignment},
		rule.Registration{Rule: NoUnnecessaryTypeConstraint},
		rule.Registration{Rule: NoUnsafeDeclarationMerging},
		rule.Registration{Rule: NoUnsafeFunctionType},
		rule.Registration{Rule: NoUnsafeUnaryMinus},
		rule.Registration{Rule: NoUselessEmptyExport},
		rule.Registration{Rule: NoWrapperObjectTypes},
		rule.Registration{Rule: PreferAsConst},
		rule.Registration{Rule: PreferNamespaceKeyword},
		rule.Registration{
			Rule:   TripleSlashReference,
			Decode: DecodeTripleSlashReferenceOptions,
		},
		rule.Registration{
			Rule:   SwitchExhaustivenessCheck,
			Decode: rule.DecodeOptionsInto[SwitchExhaustivenessCheckOptions](),
		},
	)
}
