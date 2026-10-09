// Package structure holds the rules for this project's own conventions, as opposed to the language
// rules in core and the cross-cutting ones in nexus.
package structure

import "strings"

// FileContext classifies a filename, which is enough for several rules to decide whether they apply
// before a single node is visited.
//
// The TypeScript original computes eighteen flags and memoizes them in a module-level map. Only the
// four below have a consumer here, and the rest are ported when a rule needs them: a flag nobody
// reads is a claim nobody checks, and this file is the wrong place to keep a shadow copy of the
// original's whole surface.
type FileContext struct {
	// IsReactFile is a .tsx or .jsx extension. JSX is what makes a component file, so a rule about
	// components declines everything else.
	IsReactFile bool

	// IsSpecialNextJsFile marks the files Next.js reads by name rather than by import. Their shape
	// is the framework's contract rather than an authoring choice, so a rule about how many
	// components belong in a file has no standing in one.
	IsSpecialNextJsFile bool

	// IsNetworkServiceFile marks the one implementation allowed to touch the raw fetch primitive.
	// Every rule routing network access through NetworkService has to exempt NetworkService itself,
	// or the rule forbids the thing it is asking people to use.
	IsNetworkServiceFile bool

	// IsLinkComponentFile marks the Link component's own implementation, which is the one place an
	// <a> element is allowed. Same shape as IsNetworkServiceFile: a rule saying "use Link instead"
	// has to exempt the file where Link is built, or it forbids the thing it recommends.
	IsLinkComponentFile bool

	// IsHorizontalRuleComponentFile is the same exemption for <hr> and the HorizontalRule component.
	IsHorizontalRuleComponentFile bool

	// IsInLibrariesStructure marks a file inside the shared component library.
	//
	// The direction of the boundary is what makes this different from the exemptions above. Those
	// name one file that is allowed to do what a rule forbids everywhere else; this names the side
	// of the tree where a stricter rule applies. Library code may only use values the library itself
	// defines, because a project-specific value compiled into a shared component is a dependency
	// pointing the wrong way: the library would need the project to build.
	IsInLibrariesStructure bool

	// IsPageFile marks a Next.js page, whose default export is the framework's contract rather
	// than an authoring choice: Next resolves the route by that export and by nothing else.
	IsPageFile bool

	// IsLayoutFile marks a Next.js layout, which shares the page's framework-named API surface.
	IsLayoutFile bool

	// IsLocalStorageServiceFile marks the storage service and its internal utilities, the one place
	// allowed to reach raw localStorage. Same shape as the two above, and it covers the internal
	// directory as well as the entry point, because the parsing and quota handling that justify the
	// service live there rather than in the file that re-exports them.
	IsLocalStorageServiceFile bool
}

// nextJsSpecialFileBaseNames are the files Next.js resolves by name. Both React extensions apply to
// each, and route additionally exists as a plain .ts API route.
var nextJsSpecialFileBaseNames = []string{
	"page", "layout", "error", "global-error", "not-found", "loading",
}

// FileContextFor classifies one path.
//
// Backslashes are folded to forward slashes first, so a Windows-shaped path classifies the same way.
// The original does this too, and it matters here: every check below is a suffix or substring test
// that a backslash would silently defeat.
func FileContextFor(fileName string) FileContext {
	normalizedPath := strings.ReplaceAll(fileName, `\`, "/")

	isReactFile := strings.HasSuffix(normalizedPath, ".tsx") || strings.HasSuffix(normalizedPath, ".jsx")

	isSpecialNextJsFile := strings.HasSuffix(normalizedPath, "/route.ts") ||
		strings.HasSuffix(normalizedPath, "/route.tsx")
	for _, baseName := range nextJsSpecialFileBaseNames {
		if strings.HasSuffix(normalizedPath, "/"+baseName+".tsx") ||
			strings.HasSuffix(normalizedPath, "/"+baseName+".jsx") {
			isSpecialNextJsFile = true
			break
		}
	}

	return FileContext{
		IsReactFile:         isReactFile,
		IsSpecialNextJsFile: isSpecialNextJsFile,
		// Every file of the NetworkService module, by where it lives rather than by name, the same way
		// Link's exemption is matched: the module is split across an internal/ directory whose files
		// each own part of what the rules forbid elsewhere, and a file elsewhere named after
		// NetworkService must not exempt itself.
		IsNetworkServiceFile: strings.Contains(normalizedPath, "/source/services/network/"),

		// Matched on the full path rather than the base name, which is the original's choice and
		// the load-bearing half: any file named Link.tsx would otherwise exempt itself from a rule
		// about <a> elements simply by being named after the component it is supposed to use.
		IsLinkComponentFile: strings.Contains(normalizedPath, "/components/navigation/Link.tsx") ||
			strings.Contains(normalizedPath, "/components/navigation/Link.jsx"),

		IsHorizontalRuleComponentFile: strings.Contains(normalizedPath, "/components/layout/HorizontalRule.tsx") ||
			strings.Contains(normalizedPath, "/components/layout/HorizontalRule.jsx"),

		// A substring rather than a prefix, matching the original. The library is a submodule and
		// sits at a different depth depending on which project checked it out, so anchoring at the
		// repository root would make the boundary depend on the consumer's layout.
		IsInLibrariesStructure: strings.Contains(normalizedPath, "/libraries/structure/"),

		IsPageFile: strings.HasSuffix(normalizedPath, "/page.tsx") ||
			strings.HasSuffix(normalizedPath, "/page.jsx"),

		IsLayoutFile: strings.HasSuffix(normalizedPath, "/layout.tsx") ||
			strings.HasSuffix(normalizedPath, "/layout.jsx"),

		IsLocalStorageServiceFile: strings.Contains(normalizedPath, "/services/local-storage/LocalStorageService.ts") ||
			strings.Contains(normalizedPath, "/services/local-storage/internal/LocalStorageServiceUtilities.ts"),
	}
}
