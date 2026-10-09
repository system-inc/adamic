package structure

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestConsistencyOrganizeImportsClassification is the differential against the TypeScript rule.
//
// Each row is one import statement and the group the LIVE original put it in, harvested by running
// that rule's own fixer through ESLint's Linter API and reading back the header it emitted. No row
// here was reasoned about: the table is a recording, and it is the only thing standing between this
// port and a classifier that agrees with its own author.
//
// The verdict is read the same way round, by asking this rule whether a file holding that one import
// under that one header is organized. A header naming the right group is silent and any other header
// reports, so a disagreement about classification surfaces as a finding rather than as silence.
func TestConsistencyOrganizeImportsClassification(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		importLine string
		group      string
	}{
		{importLine: "import x from 'react';", group: "Frameworks"},
		{importLine: "import x from 'react-dom';", group: "Frameworks"},
		{importLine: "import x from 'next';", group: "Frameworks"},
		{importLine: "import x from 'next/link';", group: "Frameworks"},
		{importLine: "import x from 'react/jsx-runtime';", group: "Frameworks"},
		{importLine: "import x from 'react-dom/client';", group: "Frameworks"},
		{importLine: "import x from 'next-intl';", group: "Third-party"},
		{importLine: "import x from 'lodash';", group: "Third-party"},
		{importLine: "import x from 'node:path';", group: "Node"},
		{importLine: "import x from 'node:fs';", group: "Node"},
		{importLine: "import x from 'fs';", group: "Node"},
		{importLine: "import x from 'path';", group: "Node"},
		{importLine: "import x from 'crypto';", group: "Node"},
		{importLine: "import x from 'fs/promises';", group: "Node"},
		{importLine: "import x from 'zlib';", group: "Node"},
		{importLine: "import x from '@nexus/source/Thing';", group: "Nexus"},
		{importLine: "import x from './libraries/nexus/Thing';", group: "Nexus"},
		{importLine: "import x from '@structure/source/nexus/Thing';", group: "Nexus"},
		{importLine: "import x from '@structure/assets/logo.svg';", group: "Assets"},
		{importLine: "import x from './logo.svg';", group: "Assets"},
		{importLine: "import x from '@phosphor-icons/react';", group: "Assets"},
		{importLine: "import x from 'motion';", group: "Animations"},
		{importLine: "import x from 'motion/react';", group: "Animations"},
		{importLine: "import x from 'motionx';", group: "Animations"},
		{importLine: "import x from '@structure/source/appearance/Thing';", group: "Appearance"},
		{importLine: "import x from '@project/appearance/Thing';", group: "Appearance"},
		{importLine: "import x from '@structure/source/theme/Thing';", group: "Theme"},
		{importLine: "import x from '@project/theme/Thing';", group: "Theme"},
		{importLine: "import x from '@structure/source/localization/Thing';", group: "Localization"},
		{importLine: "import x from '@project/localization/Thing';", group: "Localization"},
		{importLine: "import x from '@structure/source/_localization/Thing';", group: "Localization"},
		{importLine: "import x from '@project/_localization/Thing';", group: "Localization"},
		{importLine: "import x from '@structure/source/locales/Thing';", group: "Localization"},
		{importLine: "import x from '@project/locales/Thing';", group: "Localization"},
		{importLine: "import x from '@structure/source/context/Thing';", group: "Context"},
		{importLine: "import x from '@project/context/Thing';", group: "Context"},
		{importLine: "import x from '@structure/source/services/Thing';", group: "Services"},
		{importLine: "import x from '@project/services/Thing';", group: "Services"},
		{importLine: "import x from '@structure/source/data/Thing';", group: "Data"},
		{importLine: "import x from '@project/data/Thing';", group: "Data"},
		{importLine: "import x from '@structure/source/shared-state/Thing';", group: "Shared State"},
		{importLine: "import x from '@project/shared-state/Thing';", group: "Shared State"},
		{importLine: "import x from '@structure/source/api/Thing';", group: "APIs"},
		{importLine: "import x from '@project/api/Thing';", group: "APIs"},
		{importLine: "import x from '@structure/source/hooks/Thing';", group: "Hooks"},
		{importLine: "import x from '@project/hooks/Thing';", group: "Hooks"},
		{importLine: "import x from '@structure/source/animations/Thing';", group: "Animations"},
		{importLine: "import x from '@project/animations/Thing';", group: "Animations"},
		{importLine: "import x from '@structure/source/components/Thing';", group: "Components"},
		{importLine: "import x from '@project/components/Thing';", group: "Components"},
		{importLine: "import x from '@structure/source/modules/Thing';", group: "Components"},
		{importLine: "import x from '@project/modules/Thing';", group: "Components"},
		{importLine: "import x from '@structure/source/layouts/Thing';", group: "Layouts"},
		{importLine: "import x from '@project/layouts/Thing';", group: "Layouts"},
		{importLine: "import x from '@structure/source/pages/Thing';", group: "Components"},
		{importLine: "import x from '@project/pages/Thing';", group: "Components"},
		{importLine: "import x from '@structure/source/utilities/Thing';", group: "Utilities"},
		{importLine: "import x from '@project/utilities/Thing';", group: "Utilities"},
		{importLine: "import x from '@structure/source/_translations/Thing';", group: "Localization"},
		{importLine: "import x from '@structure/Thing';", group: "Components"},
		{importLine: "import x from '@project/Thing';", group: "Components"},
		{importLine: "import x from '@project/ProjectSettings';", group: "Project"},
		{importLine: "import x from '@structure/source/app/Thing';", group: "Components"},
		{importLine: "import x from '@structure/source/app';", group: "Components"},
		{importLine: "import x from '@structure/source/api/hooks/Thing';", group: "APIs"},
		{importLine: "import x from '@structure/source/hooks/api/Thing';", group: "APIs"},
		{importLine: "import x from '@structure/source/Constants';", group: "Constants"},
		{importLine: "import x from '@structure/source/ThingConfig';", group: "Constants"},
		{importLine: "import x from '@structure/source/Enums';", group: "Constants"},
		{importLine: "import x from '@structure/source/ThingProvider';", group: "Providers"},
		{importLine: "import x from '@structure/source/ThingProviders';", group: "Providers"},
		{importLine: "import x from '@structure/source/ProviderThing';", group: "Components"},
		{importLine: "import x from '@structure/source/ConfigThing';", group: "Components"},
		{importLine: "import x from './Constants';", group: "Constants"},
		{importLine: "import x from './ThingConfig';", group: "Constants"},
		{importLine: "import x from './ThingProvider';", group: "Providers"},
		{importLine: "import { useThing } from '@structure/source/components/Button';", group: "Hooks"},
		{importLine: "import { ThingProvider } from '@structure/source/components/Button';", group: "Providers"},
		{importLine: "import { useThing } from './Thing';", group: "Hooks"},
		{importLine: "import { ThingProvider } from './Thing';", group: "Providers"},
		{importLine: "import { useThing } from 'lodash';", group: "Hooks"},
		{importLine: "import { Thing } from 'lodash';", group: "Third-party"},
		{importLine: "import { useThing } from '@structure/source/hooks/x';", group: "Hooks"},
		{importLine: "import { useThing } from '@structure/Thing';", group: "Hooks"},
		{importLine: "import { useThing } from '@structure/source/app/Thing';", group: "Hooks"},
		{importLine: "import { ThingProvider } from '@structure/source/app/Thing';", group: "Providers"},
		{importLine: "import { usething } from './Thing';", group: "Local Components"},
		{importLine: "import { use } from './Thing';", group: "Local Components"},
		{importLine: "import { Provider } from './Thing';", group: "Providers"},
		{importLine: "import { Providers } from './Thing';", group: "Providers"},
		{importLine: "import { MyProviders } from './Thing';", group: "Providers"},
		{importLine: "import { Thing as useThing } from './Thing';", group: "Hooks"},
		{importLine: "import { useThing as Thing } from './Thing';", group: "Local Components"},
		{importLine: "import useThing from './Thing';", group: "Hooks"},
		{importLine: "import * as useThing from './Thing';", group: "Hooks"},
		{importLine: "import ThingProvider from './Thing';", group: "Providers"},
		{importLine: "import type { T } from '@structure/source/components/Button';", group: "Types"},
		{importLine: "import type { T } from '@structure/source/hooks/useThing';", group: "Types"},
		{importLine: "import type { T } from './Thing';", group: "Types"},
		{importLine: "import type { T } from './hooks/useThing';", group: "Hooks"},
		{importLine: "import type { T } from 'lodash';", group: "Types"},
		{importLine: "import type { T } from 'react';", group: "Frameworks"},
		{importLine: "import type { T } from 'node:path';", group: "Node"},
		{importLine: "import type { T } from '@nexus/source/Thing';", group: "Nexus"},
		{importLine: "import type { T } from './logo.svg';", group: "Assets"},
		{importLine: "import type { T } from '@structure/source/Constants';", group: "Types"},
		{importLine: "import { type T } from '@structure/source/components/Button';", group: "Components"},
		{importLine: "import x from './Thing';", group: "Local Components"},
		{importLine: "import x from '../Thing';", group: "Local Components"},
		{importLine: "import x from '../../Thing';", group: "Local Components"},
		{importLine: "import x from './utilities/Strings';", group: "Utilities"},
		{importLine: "import x from './api/Api';", group: "APIs"},
		{importLine: "import x from './hooks/useThing';", group: "Hooks"},
		{importLine: "import x from './useThing';", group: "Hooks"},
		{importLine: "import x from './deep/useThing';", group: "Hooks"},
		{importLine: "import x from './data/D';", group: "Data"},
		{importLine: "import x from './appearance/A';", group: "Appearance"},
		{importLine: "import x from './theme/T';", group: "Theme"},
		{importLine: "import x from './layouts/L';", group: "Layouts"},
		{importLine: "import x from './context/C';", group: "Context"},
		{importLine: "import x from './MyContext';", group: "Context"},
		{importLine: "import x from './ContextThing';", group: "Context"},
		{importLine: "import x from './services/S';", group: "Services"},
		{importLine: "import x from './shared-state/SS';", group: "Shared State"},
		{importLine: "import x from './animations/An';", group: "Animations"},
		{importLine: "import x from './components/Button';", group: "Local Components"},
		{importLine: "import x from './modules/M';", group: "Local Components"},
		{importLine: "import x from './pages/P';", group: "Local Components"},
		{importLine: "import 'alpha';", group: "Third-party"},
		{importLine: "import '@structure/source/utilities/A';", group: "Utilities"},
		{importLine: "import x from 'some-package/logo.svg';", group: "Assets"},
		{importLine: "import x from '@structure/source/icons/logo.svg';", group: "Assets"},
		{importLine: "import x from '@project/logo.svg';", group: "Assets"},
		{importLine: "import { useThing } from '@structure/source/utilities/Strings';", group: "Utilities"},
		{importLine: "import { useThing } from '@structure/source/theme/Theme';", group: "Theme"},
		{importLine: "import { useThing } from '@structure/source/layouts/Main';", group: "Layouts"},
		{importLine: "import { useThing } from '@structure/source/data/D';", group: "Data"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.importLine+" -> "+testCase.group, func(t *testing.T) {
			t.Parallel()
			source := strings.Join([]string{
				"// Dependencies - " + testCase.group,
				testCase.importLine,
				"",
				"export const X = 1;",
				"",
			}, "\n")
			result := rule_testing.Run(t, ConsistencyRequireOrganizedImports, "/repository/app/Thing.tsx", source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
