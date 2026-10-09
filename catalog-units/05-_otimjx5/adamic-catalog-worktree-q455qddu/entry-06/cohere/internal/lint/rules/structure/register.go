package structure

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
		rule.Registration{Rule: BoundaryNoProjectThemeValue},
		rule.Registration{Rule: ConsistencyRequireOrganizedImports},
		rule.Registration{Rule: NetworkNoDirectFetch},
		rule.Registration{Rule: NetworkNoForbiddenImport},
		rule.Registration{Rule: NetworkNoInvalidateCacheInOnSuccess},
		rule.Registration{Rule: NetworkNoInvalidateCacheLiteralKey},
		rule.Registration{Rule: NetworkNoStringLiteralQuery},
		rule.Registration{Rule: NetworkRequireHookOptionsParameter},
		rule.Registration{Rule: NetworkRequireHookRequestSuffix},
		rule.Registration{Rule: NetworkRequireHookVariablesType},
		rule.Registration{Rule: NextNoPageState},
		rule.Registration{Rule: NextRequireApiParameterName},
		rule.Registration{Rule: NextRequirePageDefaultExport},
		rule.Registration{Rule: ReactComponentNoConstAssignment},
		rule.Registration{Rule: ReactComponentNoDestructuring},
		rule.Registration{Rule: ReactComponentNoDisplayName},
		rule.Registration{Rule: ReactComponentNoForwardRef},
		rule.Registration{Rule: ReactComponentNoSeparateNamedExport},
		rule.Registration{Rule: ReactComponentRequireNamedExport},
		rule.Registration{Rule: ReactComponentRequirePropertiesParameter},
		rule.Registration{Rule: ReactComponentRequirePropertiesTypeSuffix},
		rule.Registration{Rule: ReactHookNoAnyType},
		rule.Registration{Rule: ReactHookNoDestructuring},
		rule.Registration{Rule: ReactHookNoPropertiesInDependencies},
		rule.Registration{Rule: ReactHookRequireEffectComment},
		rule.Registration{Rule: ReactHookRequireResultNaming},
		rule.Registration{Rule: ImportRequireReactNamespace},
		rule.Registration{Rule: ReactElementNoAnchor},
		rule.Registration{Rule: ReactElementNoHorizontalRule},
		rule.Registration{Rule: StorageNoDirectLocalStorage},
	)
}
