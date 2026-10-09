package next

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
		rule.Registration{Rule: GoogleFontDisplay},
		rule.Registration{Rule: GoogleFontPreconnect},
		rule.Registration{Rule: InlineScriptId},
		rule.Registration{Rule: NextScriptForGa},
		rule.Registration{Rule: NoAssignModuleVariable},
		rule.Registration{Rule: NoAsyncClientComponent},
		rule.Registration{Rule: NoBeforeInteractiveScriptOutsideDocument},
		rule.Registration{Rule: NoCssTags},
		rule.Registration{Rule: NoDuplicateHead},
		rule.Registration{Rule: NoDocumentImportInPage},
		rule.Registration{Rule: NoHeadElement},
		// DecodeAt because its pages directories are paths, anchored to the config's directory.
		rule.Registration{Rule: NoHtmlLinkForPages, DecodeAt: decodeNoHtmlLinkForPagesOptions},
		rule.Registration{Rule: NoHeadImportInDocument},
		rule.Registration{Rule: NoImgElement},
		rule.Registration{Rule: NoLocationAssignRelativeDestination},
		rule.Registration{Rule: NoStyledJsxInDocument},
		rule.Registration{Rule: NoPageCustomFont},
		rule.Registration{Rule: NoScriptComponentInHead},
		rule.Registration{Rule: NoSyncScripts},
		rule.Registration{Rule: NoTypos},
		rule.Registration{Rule: NoTitleInDocumentHead},
		rule.Registration{Rule: NoUnwantedPolyfillio},
	)
}
