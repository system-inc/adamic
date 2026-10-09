package module_roots

import "testing"

// TestSetSparesFrameworkFiles is the half that proves the root set declines. Every path here is
// alive by convention with no import statement anywhere in the tree, and reporting one would be the
// failure that gets the unused report closed and never reopened.
func TestSetSparesFrameworkFiles(t *testing.T) {
	t.Parallel()
	roots := &Set{dynamicImportDirectories: []string{"/translations/"}}

	for _, testCase := range []struct {
		name string
		path string
		want Reason
	}{
		{"app router page", "/repo/app/os/page.tsx", ReasonNextAppRoute},
		{"app router route handler", "/repo/app/api/tasks/route.ts", ReasonNextAppRoute},
		{"app router layout", "/repo/app/layout.tsx", ReasonNextAppRoute},
		{"app router error", "/repo/app/error.tsx", ReasonNextAppRoute},
		{"app router not-found", "/repo/app/not-found.tsx", ReasonNextAppRoute},
		{"nested project app router", "/repo/projects/www-ahra-ai/app/page.tsx", ReasonNextAppRoute},
		{"pages directory route", "/repo/pages/about.tsx", ReasonNextPagesRoute},
		{"next config", "/repo/next.config.ts", ReasonBuildConfig},
		{"open next config", "/repo/open-next.config.ts", ReasonBuildConfig},
		{"instrumentation", "/repo/instrumentation.ts", ReasonBuildConfig},
		{"middleware", "/repo/middleware.ts", ReasonBuildConfig},
		{"project settings", "/repo/ProjectSettings.tsx", ReasonBuildConfig},
		{"tailwind configuration", "/repo/TailwindConfiguration.ts", ReasonBuildConfig},
		{"declaration file", "/repo/next-env.d.ts", ReasonDeclaration},
		{"global declarations", "/repo/libraries/structure/GlobalDeclarations.d.ts", ReasonDeclaration},
		{"test file", "/repo/modules/Thing.test.ts", ReasonTest},
		{"spec file", "/repo/modules/Thing.spec.tsx", ReasonTest},
		{"dynamically imported locale", "/repo/libraries/structure/source/localization/translations/de.ts", ReasonDynamicImportTarget},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			isRoot, reason := roots.IsRoot(testCase.path)
			if !isRoot {
				t.Fatalf("%s was not spared, so a live file would be reported as unused", testCase.path)
			}
			if reason != testCase.want {
				t.Fatalf("spared for %q, want %q", reason, testCase.want)
			}
		})
	}
}

// TestSetDoesNotSpareOrdinaryFiles is the half that proves the root set can still report. A root
// set that spares everything is indistinguishable from a broken analysis, and it produces exactly
// the clean-looking empty report this project exists to make impossible.
func TestSetDoesNotSpareOrdinaryFiles(t *testing.T) {
	t.Parallel()
	roots := &Set{dynamicImportDirectories: []string{"/translations/"}}

	for _, path := range []string{
		"/repo/libraries/structure/source/utilities/Strings.ts",
		"/repo/app/(os-layout)/_components/profile/OsProfilePage.tsx",
		"/repo/modules/tasks/TasksCommandLineInterface.ts",
		// A private component parked under pages/ is not a route, so it stays reportable.
		"/repo/pages/_components/Widget.tsx",
		// Named like a route but not in a route directory.
		"/repo/modules/errors/error.ts",
		// A near-miss on the marker word, which must not be read as a convention.
		"/repo/modules/apple/Pages.ts",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if isRoot, reason := roots.IsRoot(path); isRoot {
				t.Fatalf("%s was spared as %q, so a genuine finding would be hidden", path, reason)
			}
		})
	}
}
