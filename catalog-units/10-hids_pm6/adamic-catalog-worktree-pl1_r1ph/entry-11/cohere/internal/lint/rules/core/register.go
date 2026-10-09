package core

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
		rule.Registration{Rule: ConstructorSuper},
		rule.Registration{Rule: ForDirection},
		rule.Registration{
			Rule:   GetterReturn,
			Decode: rule.DecodeOptionsInto[GetterReturnOptions](),
		},
		rule.Registration{Rule: NoArrayConstructor},
		rule.Registration{Rule: NoAsyncPromiseExecutor},
		rule.Registration{Rule: NoCaseDeclarations},
		rule.Registration{Rule: NoClassAssign},
		rule.Registration{Rule: NoConstAssign},
		rule.Registration{Rule: NoCompareNegZero},
		rule.Registration{
			Rule:   NoCondAssign,
			Decode: rule.DecodeOptionsInto[NoCondAssignOptions](),
		},
		rule.Registration{Rule: NoCaller},
		rule.Registration{Rule: NoControlRegex},
		rule.Registration{Rule: NoDupeClassMembers},
		rule.Registration{
			Rule:   NoFallthrough,
			Decode: rule.DecodeOptionsInto[NoFallthroughOptions](),
		},
		rule.Registration{Rule: NoImportAssign},
		rule.Registration{
			Rule:   NoIrregularWhitespace,
			Decode: rule.DecodeOptionsInto[NoIrregularWhitespaceOptions](),
		},
		rule.Registration{Rule: NoIterator},
		rule.Registration{Rule: NoLossOfPrecision},
		rule.Registration{Rule: NoNewNativeNonconstructor},
		rule.Registration{Rule: NoObjCalls},
		rule.Registration{Rule: NoOctal},
		rule.Registration{Rule: NoPrototypeBuiltins},
		rule.Registration{
			Rule:   NoUnusedExpressions,
			Decode: rule.DecodeOptionsInto[NoUnusedExpressionsOptions](),
		},
		rule.Registration{
			Rule:   NoUnusedVars,
			Decode: rule.DecodeOptionsInto[NoUnusedVarsOptions](),
		},
		rule.Registration{Rule: NoUnusedPrivateClassMembers},
		rule.Registration{
			Rule:   PreferConst,
			Decode: rule.DecodeOptionsInto[PreferConstOptions](),
		},
		rule.Registration{Rule: PreferRestParams},
		rule.Registration{
			Rule:   NoConstantBinaryExpression,
			Decode: rule.DecodeOptionsInto[NoConstantBinaryExpressionOptions](),
		},
		rule.Registration{
			Rule:   NoConstantCondition,
			Decode: rule.DecodeOptionsInto[NoConstantConditionOptions](),
		},
		rule.Registration{
			Rule:   NoExtraBooleanCast,
			Decode: rule.DecodeOptionsInto[NoExtraBooleanCastOptions](),
		},
		rule.Registration{Rule: NoDebugger},
		rule.Registration{Rule: NoDeleteVar},
		rule.Registration{Rule: NoDupeElseIf},
		rule.Registration{Rule: NoDuplicateCase},
		rule.Registration{
			Rule:   NoEmpty,
			Decode: rule.DecodeOptionsInto[NoEmptyOptions](),
		},
		rule.Registration{Rule: NoEmptyCharacterClass},
		rule.Registration{
			Rule:   NoEmptyPattern,
			Decode: rule.DecodeOptionsInto[NoEmptyPatternOptions](),
		},
		rule.Registration{Rule: NoEmptyStaticBlock},
		rule.Registration{
			Rule:   NoEval,
			Decode: rule.DecodeOptionsInto[NoEvalOptions](),
		},
		rule.Registration{Rule: NoExAssign},
		rule.Registration{Rule: NoFuncAssign},
		rule.Registration{
			Rule:   NoGlobalAssign,
			Decode: rule.DecodeOptionsInto[NoGlobalAssignOptions](),
		},
		rule.Registration{
			Rule:   NoInvalidRegexp,
			Decode: DecodeNoInvalidRegexpOptions,
		},
		rule.Registration{
			Rule:   NoMisleadingCharacterClass,
			Decode: rule.DecodeOptionsInto[NoMisleadingCharacterClassOptions](),
		},
		rule.Registration{Rule: NoNonoctalDecimalEscape},
		rule.Registration{Rule: NoRegexSpaces},
		rule.Registration{Rule: NoSetterReturn},
		rule.Registration{
			Rule:   NoShadowRestrictedNames,
			Decode: rule.DecodeOptionsInto[NoShadowRestrictedNamesOptions](),
		},
		rule.Registration{Rule: NoSparseArrays},
		rule.Registration{Rule: NoThisBeforeSuper},
		rule.Registration{Rule: NoUnassignedVars},
		rule.Registration{Rule: NoUselessAssignment},
		rule.Registration{Rule: NoUnsafeFinally},
		rule.Registration{
			Rule:   NoUnsafeNegation,
			Decode: rule.DecodeOptionsInto[NoUnsafeNegationOptions](),
		},
		rule.Registration{
			Rule:   NoUnsafeOptionalChaining,
			Decode: rule.DecodeOptionsInto[NoUnsafeOptionalChainingOptions](),
		},
		rule.Registration{Rule: NoUnusedLabels},
		rule.Registration{Rule: NoUselessBackreference},
		rule.Registration{Rule: NoUselessCatch},
		rule.Registration{
			Rule:   NoUselessEscape,
			Decode: rule.DecodeOptionsInto[NoUselessEscapeOptions](),
		},
		rule.Registration{Rule: NoVar},
		rule.Registration{Rule: NoWith},
		rule.Registration{Rule: PreferSpread},
		rule.Registration{
			Rule:   PreserveCaughtError,
			Decode: rule.DecodeOptionsInto[PreserveCaughtErrorOptions](),
		},
		rule.Registration{Rule: RequireYield},
		rule.Registration{
			Rule:   UseIsNaN,
			Decode: rule.DecodeOptionsInto[UseIsNaNOptions](),
		},
		rule.Registration{
			Rule:   ValidTypeof,
			Decode: rule.DecodeOptionsInto[ValidTypeofOptions](),
		},
	)
}
