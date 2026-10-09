package structure

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// A hook file named in Adamic is named by the same rule as a `.ts` one: `User.a` lacks the Request
// suffix and `UserRequest.a` has it (#kwt1htp).
func TestNetworkRequireHookRequestSuffixNamesAnAdamicFileAsTypeScript(t *testing.T) {
	t.Parallel()
	source := requestSuffixDeclarations + "export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n"
	for _, fileName := range []string{plainNamedFile, requestSuffixFile} {
		rule_testing.ExpectSameFindings(t,
			rule_testing.Run(t, NetworkRequireHookRequestSuffix, fileName, source),
			rule_testing.Run(t, NetworkRequireHookRequestSuffix, strings.TrimSuffix(fileName, ".ts")+".a", source))
	}
}

// An Adamic `ButtonTheme.a` is a theme file as `ButtonTheme.ts` is, so its values bound a component's
// props the same way (#kwt1htp).
func TestBoundaryNoProjectThemeValueReadsAnAdamicThemeFile(t *testing.T) {
	t.Parallel()
	source := "export const a = <Button variant=\"Emphasized\" />;\n"
	rule_testing.ExpectSameFindings(t,
		rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			themeFilePath:   buttonThemeSource,
			libraryFilePath: source,
		}, libraryFilePath),
		rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			strings.TrimSuffix(themeFilePath, ".ts") + ".a": buttonThemeSource,
			libraryFilePath: source,
		}, libraryFilePath))
}
