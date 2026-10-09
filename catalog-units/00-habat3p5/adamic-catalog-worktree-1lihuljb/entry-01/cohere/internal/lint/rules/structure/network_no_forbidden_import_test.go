package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const forbiddenImportFile = "/repository/source/components/Thing.tsx"

func TestNetworkNoForbiddenImportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantIds    []string
	}{
		{
			"tanstack react-query", forbiddenImportFile,
			"import { useQuery } from '@tanstack/react-query';\nexport const Use = useQuery;\n",
			[]string{"noDirectTanStackQuery"},
		},
		{
			"apollo client", forbiddenImportFile,
			"import { gql } from '@apollo/client';\nexport const Use = gql;\n",
			[]string{"noDirectApollo"},
		},
		// A prefix match, so every package in the scope is caught rather than only the entry point.
		{
			"another apollo package", forbiddenImportFile,
			"import { onError } from '@apollo/client/link/error';\nexport const Use = onError;\n",
			[]string{"noDirectApollo"},
		},
		{
			"graphql from a generated path", forbiddenImportFile,
			"import { graphql } from '../generated/graphql';\nexport const Use = graphql;\n",
			[]string{"noDirectGraphqlImport"},
		},
		// The imported name is what counts rather than the local alias: the alias changes what the
		// call site says and not what was imported.
		{
			"an aliased graphql import", forbiddenImportFile,
			"import { graphql as query } from '../generated/graphql';\nexport const Use = query;\n",
			[]string{"noDirectGraphqlImport"},
		},
		{
			"tanstack in an ordinary component", "/repository/app/_components/Panel.tsx",
			"import { useQuery } from '@tanstack/react-query';\nexport const Use = useQuery;\n",
			[]string{"noDirectTanStackQuery"},
		},
		// Providers.tsx was once exempt by name; Structure's providers no longer configure TanStack,
		// so it is an ordinary file.
		{
			"tanstack inside a Providers.tsx", "/repository/app/Providers.tsx",
			"import { QueryClient } from '@tanstack/react-query';\nexport const Use = QueryClient;\n",
			[]string{"noDirectTanStackQuery"},
		},
		// A file named after NetworkService outside its module is not NetworkService.
		{
			"tanstack in a file merely named after NetworkService", "/repository/app/_services/MyNetworkService.ts",
			"import { useQuery } from '@tanstack/react-query';\nexport const Use = useQuery;\n",
			[]string{"noDirectTanStackQuery"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoForbiddenImport, testCase.fileName,
				testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNetworkNoForbiddenImportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{
			"the wrapper being used as intended", forbiddenImportFile,
			"import { useGraphQlQuery } from '@structure/source/services/network/NetworkService';\nexport const Use = useGraphQlQuery;\n",
		},
		// NetworkService's module, which is where TanStack is set up. A rule that forbade it here
		// would forbid the thing it asks people to use. The module spans internal/ too.
		{
			"tanstack inside NetworkService.ts", "/repository/source/services/network/NetworkService.ts",
			"import { QueryClient } from '@tanstack/react-query';\nexport const Use = QueryClient;\n",
		},
		{
			"tanstack inside NetworkService's provider", "/repository/source/services/network/NetworkServiceProvider.tsx",
			"import { QueryClientProvider } from '@tanstack/react-query';\nexport const Use = QueryClientProvider;\n",
		},
		{
			"tanstack inside NetworkService's internal cache", "/repository/libraries/structure/source/services/network/internal/NetworkServiceCache.ts",
			"import { QueryClient } from '@tanstack/react-query';\nexport const Use = QueryClient;\n",
		},
		// The apollo exemption is deliberately absent, so this is the boundary: the prefix must be
		// the scope rather than any occurrence of the word.
		{
			"a package merely containing the word apollo", forbiddenImportFile,
			"import { thing } from 'apollo-helpers';\nexport const Use = thing;\n",
		},
		// A generated path is an ordinary place to import types from, so the path alone is not the
		// defect. This is the case the specifier check exists for.
		{
			"types from a generated path", forbiddenImportFile,
			"import type { UserType } from '../generated/graphql';\ndeclare const user: UserType;\nexport const Use = user;\n",
		},
		{
			"a non-graphql specifier from a generated path", forbiddenImportFile,
			"import { helper } from '../generated/graphql';\nexport const Use = helper;\n",
		},
		// The word graphql outside a generated path is somebody's own module.
		{
			"graphql from a non-generated path", forbiddenImportFile,
			"import { graphql } from '@structure/source/api/GraphQlHelpers';\nexport const Use = graphql;\n",
		},
		{"no imports at all", forbiddenImportFile, "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkNoForbiddenImport, testCase.fileName,
				testCase.sourceText))
		})
	}
}
