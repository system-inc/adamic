package boundaries

import (
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * Every verdict in this file is one the installed eslint-plugin-boundaries 7.2.0 returned, driven
 * through ESLint with api-phi-health's own LintConfiguration.ts on planted imports at real paths in
 * that repository (Base d6ead90c). Upstream ships no test corpus in its package, so that run is the
 * corpus. The layouts below reproduce the paths it was run on, and the options are Base's two blocks
 * as its config writes them, with the elements moved into the options as this port reads them.
 */

// basePackageOptions is BasePackageBoundaryConfiguration: the command line may reach api and
// foundation, foundation may reach api, and anything else between them is refused.
const basePackageOptions = `{
	"default": "disallow",
	"policies": [
		{"from": {"element": {"type": "cli"}}, "allow": {"to": {"element": {"types": {"anyOf": ["api", "foundation"]}}}}},
		{"from": {"element": {"type": "foundation"}}, "allow": {"to": {"element": {"types": {"anyOf": ["api"]}}}}}
	],
	"elements": [
		{"type": "api", "pattern": "libraries/base/source/api/**"},
		{"type": "foundation", "pattern": "libraries/base/source/foundation/**"},
		{"type": "cli", "pattern": "libraries/base/command-line/**"}
	]
}`

// projectModuleMessage is the message ProjectModuleBoundaryConfiguration writes on its policy.
const projectModuleMessage = "A module cannot import from a worker. It is registered by a worker it cannot name, and the import drags that worker's whole graph into every other worker that registers this module. Move the shared code into a module, or inject it."

// projectModuleOptions is ProjectModuleBoundaryConfiguration: a module may not import a worker.
const projectModuleOptions = `{
	"default": "allow",
	"policies": [
		{"from": {"element": {"type": "project-module"}}, "disallow": {"to": {"element": {"type": "project-worker"}}}, "message": "` + projectModuleMessage + `"}
	],
	"elements": [
		{"type": "project-module", "pattern": "source/modules/*", "capture": ["moduleName"]},
		{"type": "project-source", "pattern": "source/*"},
		{"type": "project-worker", "pattern": "workers/*", "capture": ["workerName"]}
	]
}`

func decodeForTest(t *testing.T, raw string) DependenciesOptions {
	t.Helper()
	decoded, err := decodeDependenciesOptions([]byte(raw), rule.OptionsBase{})
	if err != nil {
		t.Fatalf("decoding the options: %v", err)
	}
	return decoded.(DependenciesOptions)
}

// baseLayout is the slice of Base the package-boundary cases import across.
var baseLayout = map[string]string{
	"libraries/base/source/api/rpc/client/driver/RpcClientDriver.ts": "export class RpcClientDriver {}",
	"libraries/base/source/foundation/base/Base.ts":                  "export class Base {}",
	"libraries/base/source/foundation/entity/Entity.ts":              "export class Entity {}",
	"libraries/base/source/modules/account/AccountModule.ts":         "export class AccountModule {}",
	"libraries/base/libraries/nexus/source/time/DueDate.ts":          "export class DueDate {}",
	"libraries/base/command-line/commands/CompileCommand.ts":         "export class CompileCommand {}",
	"node_modules/zod/package.json":                                  `{"name": "zod", "version": "1.0.0", "types": "index.d.ts"}`,
	"node_modules/zod/index.d.ts":                                    "export declare const z: number;",
}

// projectLayout is the slice of a Base project the module-boundary cases import across.
var projectLayout = map[string]string{
	"workers/api/ApiWorker.ts":                      "export class ApiWorker {}",
	"workers/api/index.ts":                          "export { ApiWorker } from './ApiWorker';",
	"workers/chat/ChatWorker.ts":                    "export class ChatWorker {}",
	"source/common/AnalyticsAccessRoleType.ts":      "export type AnalyticsAccessRoleType = 'Reader';",
	"source/modules/ads-platform/Ga4AdsService.ts":  "export class Ga4AdsService {}",
	"libraries/base/source/foundation/base/Base.ts": "export class Base {}",
	"vendor/workers/thing/Thing.ts":                 "export class Thing {}",
	// A package whose types sit under a `workers/<name>/` folder, so its resolved path ends in a suffix
	// the worker pattern matches. Upstream never asks: a dependency whose module origin is not local
	// is skipped before any element is described (Dependencies.js, `isLocalDependency`).
	"node_modules/worker-kit/package.json":           `{"name": "worker-kit", "version": "1.0.0", "types": "workers/api/index.d.ts"}`,
	"node_modules/worker-kit/workers/api/index.d.ts": "export declare const kit: number;",
}

func runCase(t *testing.T, layout map[string]string, subjectFile string, source string, options DependenciesOptions) rule_testing.Result {
	t.Helper()
	files := map[string]string{subjectFile: source}
	for name, contents := range layout {
		if name != subjectFile {
			files[name] = contents
		}
	}
	return rule_testing.RunTypedFilesWithOptions(t, Dependencies, files, subjectFile, options)
}

type dependencyCase struct {
	name    string
	file    string
	source  string
	wantIds []string
}

// relativeSpecifier finds the relative specifiers a fixture's source names.
var relativeSpecifier = regexp.MustCompile(`(?:from|import\(|import) '(\.[^']*)'`)

// expectEveryRelativeImportResolves refuses a clean fixture whose relative imports reach no file.
//
// A specifier that climbs one directory too many resolves to nothing, the rule skips it as upstream
// would, and the case passes having checked nothing. That happened here: "a module imports the command
// line" climbed four levels where three reach Base, and a mutant that checked files belonging to no
// element survived the whole suite because of it. A case that means to import nothing says so in its
// name.
func expectEveryRelativeImportResolves(t *testing.T, layout map[string]string, file string, source string) {
	t.Helper()
	for _, match := range relativeSpecifier.FindAllStringSubmatch(source, -1) {
		target := path.Clean(path.Join(path.Dir(file), match[1]))
		_, isFile := layout[target+".ts"]
		_, isIndex := layout[target+"/index.ts"]
		if !isFile && !isIndex {
			t.Fatalf("%q from %s reaches no file in the layout (%s), so this clean case checked nothing", match[1], file, target)
		}
	}
}

func runCases(t *testing.T, layout map[string]string, options DependenciesOptions, cases []dependencyCase) {
	t.Helper()
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := runCase(t, layout, testCase.file, testCase.source, options)
			if len(testCase.wantIds) == 0 {
				if !strings.Contains(testCase.name, "does not exist") {
					expectEveryRelativeImportResolves(t, layout, testCase.file, testCase.source)
				}
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

const apiFile = "libraries/base/source/api/rpc/client/RpcClient.ts"
const foundationFile = "libraries/base/source/foundation/base/BaseHandler.ts"

func TestDependenciesBasePackageBoundary(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, basePackageOptions)

	runCases(t, baseLayout, options, []dependencyCase{
		// api reaches nothing, in every shape upstream sees a dependency.
		{"api imports foundation", apiFile, "import { Base } from '../../../foundation/base/Base';\nexport const x = Base;", []string{"noPolicyAllows"}},
		{"api imports a foundation type", apiFile, "import type { Base } from '../../../foundation/base/Base';\nexport type X = Base;", []string{"noPolicyAllows"}},
		{"api imports the command line", apiFile, "import { CompileCommand } from '../../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;", []string{"noPolicyAllows"}},
		{"api re-exports all of foundation", apiFile, "export * from '../../../foundation/base/Base';", []string{"noPolicyAllows"}},
		{"api re-exports from foundation", apiFile, "export { Base } from '../../../foundation/base/Base';", []string{"noPolicyAllows"}},
		{"api imports foundation dynamically", apiFile, "export const x = import('../../../foundation/base/Base');", []string{"noPolicyAllows"}},
		{"api imports foundation for its side effects", apiFile, "import '../../../foundation/base/Base';", []string{"noPolicyAllows"}},

		// What upstream never checks.
		{"api imports itself", apiFile, "import { RpcClientDriver } from './driver/RpcClientDriver';\nexport const x = RpcClientDriver;", nil},
		{"api imports a module, which is no element", apiFile, "import { AccountModule } from '../../../modules/account/AccountModule';\nexport const x = AccountModule;", nil},
		{"api imports nexus, which is no element", apiFile, "import { DueDate } from '../../../../libraries/nexus/source/time/DueDate';\nexport const x = DueDate;", nil},
		{"api imports a package", apiFile, "import { z } from 'zod';\nexport const x = z;", nil},
		{"api imports a file that does not exist", apiFile, "import { Missing } from '../../../foundation/base/DoesNotExist';\nexport const x = Missing;", nil},

		// foundation may reach api and nothing else.
		{"foundation imports api", foundationFile, "import { RpcClientDriver } from '../../api/rpc/client/driver/RpcClientDriver';\nexport const x = RpcClientDriver;", nil},
		{"foundation imports the command line", foundationFile, "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;", []string{"noPolicyAllows"}},
		{"foundation imports a module", foundationFile, "import { AccountModule } from '../../modules/account/AccountModule';\nexport const x = AccountModule;", nil},
		{"foundation imports across its own folders", foundationFile, "import { Base } from './Base';\nimport { Entity } from '../entity/Entity';\nexport const x = [Base, Entity];", nil},

		// A file that is no element is never checked, whatever it imports.
		{"a module imports the command line", "libraries/base/source/modules/account/AccountService.ts", "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;", nil},
	})
}

const moduleFile = "source/modules/chat/ChatModule.ts"

func TestDependenciesProjectModuleBoundary(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, projectModuleOptions)

	runCases(t, projectLayout, options, []dependencyCase{
		{"a module imports a worker", moduleFile, "import { ApiWorker } from '../../../workers/api/ApiWorker';\nexport const x = ApiWorker;", []string{"policyMessage"}},
		{"a module imports a worker's type", moduleFile, "import type { ApiWorker } from '../../../workers/api/ApiWorker';\nexport type X = ApiWorker;", []string{"policyMessage"}},
		{"a module imports a worker's index", moduleFile, "import { ApiWorker } from '../../../workers/api/index';\nexport const x = ApiWorker;", []string{"policyMessage"}},
		{"a module imports a worker dynamically", moduleFile, "export const x = import('../../../workers/api/ApiWorker');", []string{"policyMessage"}},

		// Upstream matches a pattern against the right-hand end of a path, so a folder named `workers`
		// anywhere is a worker. Measured on the installed build with `source/*`, which claimed
		// libraries/base/source/foundation as its own element.
		{"a module imports a nested workers folder", moduleFile, "import { Thing } from '../../../vendor/workers/thing/Thing';\nexport const x = Thing;", []string{"policyMessage"}},

		{"a module imports common source", moduleFile, "import type { AnalyticsAccessRoleType } from '../../common/AnalyticsAccessRoleType';\nexport type X = AnalyticsAccessRoleType;", nil},
		{"a module imports another module", moduleFile, "import { Ga4AdsService } from '../ads-platform/Ga4AdsService';\nexport const x = Ga4AdsService;", nil},
		{"a module imports a package laid out like a worker", moduleFile, "import { kit } from 'worker-kit';\nexport const x = kit;", nil},
		{"a module imports Base", moduleFile, "import { Base } from '../../../libraries/base/source/foundation/base/Base';\nexport const x = Base;", nil},
		{"a worker imports another worker", "workers/api/Routes.ts", "import { ChatWorker } from '../chat/ChatWorker';\nexport const x = ChatWorker;", nil},
		{"common source imports a worker", "source/common/Helper.ts", "import { ApiWorker } from '../../workers/api/ApiWorker';\nexport const x = ApiWorker;", nil},
	})
}

// The custom message is the policy's own words, verbatim.
func TestDependenciesUsesThePolicyMessage(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, projectModuleOptions)
	result := runCase(t, projectLayout, moduleFile, "import { ApiWorker } from '../../../workers/api/ApiWorker';\nexport const x = ApiWorker;", options)
	rule_testing.ExpectFindings(t, result, "policyMessage")
	if result.Diagnostics[0].Message.Description != projectModuleMessage {
		t.Fatalf("expected the policy's message verbatim, got %q", result.Diagnostics[0].Message.Description)
	}
}

// A policy that matched by disallow without a message of its own still says which layers.
func TestDependenciesWordsAnUnmessagedDisallow(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, strings.Replace(projectModuleOptions, `, "message": "`+projectModuleMessage+`"`, "", 1))
	result := runCase(t, projectLayout, moduleFile, "import { ApiWorker } from '../../../workers/api/ApiWorker';\nexport const x = ApiWorker;", options)
	rule_testing.ExpectFindings(t, result, "policyDisallows")
	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, `"project-module"`) || !strings.Contains(description, `"project-worker"`) {
		t.Fatalf("expected the message to name both layers, got %q", description)
	}
}

// The last policy to match decides, and within one policy disallow is checked before allow.
func TestDependenciesLastMatchingPolicyDecides(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, `{
		"default": "disallow",
		"policies": [
			{"from": {"element": {"type": "foundation"}}, "allow": {"to": {"element": {"type": "cli"}}}},
			{"from": {"element": {"type": "foundation"}}, "disallow": {"to": {"element": {"type": "cli"}}}, "allow": {"to": {"element": {"type": "cli"}}}}
		],
		"elements": [
			{"type": "foundation", "pattern": "libraries/base/source/foundation/**"},
			{"type": "cli", "pattern": "libraries/base/command-line/**"}
		]
	}`)
	result := runCase(t, baseLayout, foundationFile, "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;", options)
	rule_testing.ExpectFindings(t, result, "policyDisallows")

	reversed := decodeForTest(t, `{
		"default": "disallow",
		"policies": [
			{"from": {"element": {"type": "foundation"}}, "disallow": {"to": {"element": {"type": "cli"}}}},
			{"from": {"element": {"type": "foundation"}}, "allow": {"to": {"element": {"type": "cli"}}}}
		],
		"elements": [
			{"type": "foundation", "pattern": "libraries/base/source/foundation/**"},
			{"type": "cli", "pattern": "libraries/base/command-line/**"}
		]
	}`)
	rule_testing.ExpectClean(t, runCase(t, baseLayout, foundationFile, "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;", reversed))
}

// An entry's own from wins over its policy's, field by field. Measured on the installed build: a policy
// from cli whose disallow names from foundation refuses foundation, and the same policy without the
// entry's from is clean.
func TestDependenciesEntrySelectorWinsOverThePolicys(t *testing.T) {
	t.Parallel()
	elements := `"elements": [
		{"type": "foundation", "pattern": "libraries/base/source/foundation/**"},
		{"type": "cli", "pattern": "libraries/base/command-line/**"}
	]`
	source := "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;"

	entryWins := decodeForTest(t, `{"default": "allow", "policies": [{"from": {"element": {"type": "cli"}}, "disallow": {"from": {"element": {"type": "foundation"}}, "to": {"element": {"type": "cli"}}}}], `+elements+`}`)
	rule_testing.ExpectFindings(t, runCase(t, baseLayout, foundationFile, source, entryWins), "policyDisallows")

	outerOnly := decodeForTest(t, `{"default": "allow", "policies": [{"from": {"element": {"type": "cli"}}, "disallow": {"to": {"element": {"type": "cli"}}}}], `+elements+`}`)
	rule_testing.ExpectClean(t, runCase(t, baseLayout, foundationFile, source, outerOnly))
}

// Within one side of a dependency, an entry's `types` replaces its policy's, the same shallow merge
// as `type`.
func TestDependenciesEntryTypesWinOverThePolicys(t *testing.T) {
	t.Parallel()
	elements := `"elements": [
		{"type": "foundation", "pattern": "libraries/base/source/foundation/**"},
		{"type": "cli", "pattern": "libraries/base/command-line/**"}
	]`
	source := "import { CompileCommand } from '../../../command-line/commands/CompileCommand';\nexport const x = CompileCommand;"
	options := decodeForTest(t, `{"default": "allow", "policies": [{"to": {"element": {"types": {"anyOf": ["foundation"]}}}, "disallow": {"to": {"element": {"types": {"anyOf": ["cli"]}}}}}], `+elements+`}`)
	rule_testing.ExpectFindings(t, runCase(t, baseLayout, foundationFile, source, options), "policyDisallows")
}

// A rule-level message words every violation no policy words itself.
func TestDependenciesUsesTheRuleMessageWhenNoPolicyHasOne(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, strings.Replace(basePackageOptions, `"default": "disallow",`, `"default": "disallow", "message": "Layers point one way.",`, 1))
	result := runCase(t, baseLayout, apiFile, "import { Base } from '../../../foundation/base/Base';\nexport const x = Base;", options)
	rule_testing.ExpectFindings(t, result, "policyMessage")
	if result.Diagnostics[0].Message.Description != "Layers point one way." {
		t.Fatalf("expected the rule's message, got %q", result.Diagnostics[0].Message.Description)
	}
}

// The finding lands on the specifier string, where upstream puts it, and not on the comment above.
func TestDependenciesReportsAtTheSpecifier(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, basePackageOptions)
	source := "// Dependencies\nimport { Base } from '../../../foundation/base/Base';\nexport const x = Base;"
	result := runCase(t, baseLayout, apiFile, source, options)
	rule_testing.ExpectFindings(t, result, "noPolicyAllows")

	text := result.SourceFile.Text()
	span := text[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if strings.TrimSpace(span) != "'../../../foundation/base/Base'" {
		t.Fatalf("expected the finding on the specifier, got %q", span)
	}
	line, _ := scanner.GetECMALineAndByteOffsetOfPosition(result.SourceFile, scanner.SkipTrivia(text, result.Diagnostics[0].Range.Pos()))
	if line != 1 {
		t.Fatalf("expected the finding on the import's line (2), got line %d", line+1)
	}
}

// The stated divergence: upstream reads a bare `require('…')`, and this port does not. Pinned so a
// later `require` branch arrives with a fixture that proves it rather than replacing this silently.
func TestDependenciesRequireInTypeScriptIsTheMeasuredDivergence(t *testing.T) {
	t.Parallel()
	options := decodeForTest(t, basePackageOptions)
	result := runCase(t, baseLayout, apiFile, "declare const require: (name: string) => unknown;\nexport const x = require('../../../foundation/base/Base');", options)
	rule_testing.ExpectClean(t, result)
}

func TestDependenciesElementMembership(t *testing.T) {
	t.Parallel()
	project := decodeForTest(t, projectModuleOptions).Elements
	base := decodeForTest(t, basePackageOptions).Elements

	cases := []struct {
		path      string
		elements  []elementDescriptor
		wantType  string
		wantPath  string
		wantFound bool
	}{
		{"source/modules/chat/ChatModule.ts", project, "project-module", "source/modules/chat", true},
		{"source/modules/chat/deep/Thing.ts", project, "project-module", "source/modules/chat", true},
		{"source/common/Thing.ts", project, "project-source", "source/common", true},
		{"source/Thing.ts", project, "", "", false},
		{"workers/api/ApiWorker.ts", project, "project-worker", "workers/api", true},
		// The right-hand matching upstream documents, and that Base's project block depends on.
		{"libraries/base/source/foundation/base/Base.ts", project, "project-source", "libraries/base/source/foundation", true},
		{"libraries/base/source/modules/account/AccountModule.ts", project, "project-module", "libraries/base/source/modules/account", true},
		// micromatch's default keeps `*` and `**` out of a dot segment.
		{"workers/api/.wrangler/state.ts", project, "", "", false},
		{"workers/.hidden/state.ts", project, "", "", false},
		// The shortest matching suffix wins, so a workers folder inside a module is a worker.
		{"source/modules/chat/workers/api/Thing.ts", project, "project-worker", "source/modules/chat/workers/api", true},
		{"libraries/base/source/api/rpc/client/RpcClient.ts", base, "api", "libraries/base/source/api", true},
		{"libraries/base/source/foundation/base/Base.ts", base, "foundation", "libraries/base/source/foundation", true},
		{"libraries/base/command-line/base-cli.ts", base, "cli", "libraries/base/command-line", true},
		{"libraries/base/source/modules/account/AccountModule.ts", base, "", "", false},
	}
	for _, testCase := range cases {
		got, found := describeElement(testCase.path, testCase.elements)
		if found != testCase.wantFound || got.Type != testCase.wantType || got.Path != testCase.wantPath {
			t.Errorf("%s: expected (%q, %q, %v), got (%q, %q, %v)", testCase.path,
				testCase.wantType, testCase.wantPath, testCase.wantFound, got.Type, got.Path, found)
		}
	}
}

// Every shape this port does not decide is refused by name rather than loaded and half applied.
func TestDependenciesRefusesWhatItDoesNotPort(t *testing.T) {
	t.Parallel()
	elements := `"elements": [{"type": "api", "pattern": "api/**"}]`
	cases := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"no options", ``, "needs its options"},
		{"no elements", `{"default": "disallow"}`, "has no elements"},
		{"a legacy string selector", `{"policies": [{"from": "api", "allow": "api"}], ` + elements + `}`, "legacy string selectors"},
		{"a captured selector", `{"policies": [{"from": {"element": {"type": "api", "captured": {"name": "x"}}}, "allow": {"to": {"element": {"type": "api"}}}}], ` + elements + `}`, "captured"},
		{"a file selector", `{"policies": [{"from": {"file": {"path": "x"}}, "allow": {"to": {"element": {"type": "api"}}}}], ` + elements + `}`, "file"},
		{"an allOf query", `{"policies": [{"from": {"element": {"types": {"allOf": ["api"]}}}, "allow": {"to": {"element": {"type": "api"}}}}], ` + elements + `}`, "allOf"},
		{"a type pattern", `{"policies": [{"from": {"element": {"type": "api*"}}, "allow": {"to": {"element": {"type": "api"}}}}], ` + elements + `}`, "micromatch pattern"},
		{"a templated message", `{"message": "{{from.type}} may not", ` + elements + `}`, "template"},
		{"the legacy rules key", `{"rules": [], ` + elements + `}`, "rules"},
		{"checkInternals true", `{"checkInternals": true, ` + elements + `}`, "checkInternals"},
		{"file mode", `{"elements": [{"type": "api", "pattern": "api/**", "mode": "file"}]}`, "folder mode"},
		{"a brace pattern", `{"elements": [{"type": "api", "pattern": "{api,client}/**"}]}`, "brace or extglob"},
		{"a basePattern", `{"elements": [{"type": "api", "pattern": "api/**", "basePattern": "x"}]}`, "basePattern"},
		{"a dependency selector", `{"policies": [{"dependency": {"kind": "type"}, "allow": {"to": {"element": {"type": "api"}}}}], ` + elements + `}`, "dependency"},
		{"a policy that decides nothing", `{"policies": [{"from": {"element": {"type": "api"}}}], ` + elements + `}`, "neither allow nor disallow"},
		{"an unknown default", `{"default": "deny", ` + elements + `}`, "default"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeDependenciesOptions([]byte(testCase.raw), rule.OptionsBase{})
			if err == nil {
				t.Fatalf("expected a refusal mentioning %q, the options loaded", testCase.wantErr)
			}
			if !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("expected a refusal mentioning %q, got %v", testCase.wantErr, err)
			}
		})
	}
}

// What Base's config writes loads, and the root is anchored where the config was written.
func TestDependenciesDecodesBasesOwnBlocksAndAnchorsTheRoot(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{basePackageOptions, projectModuleOptions} {
		decoded, err := decodeDependenciesOptions([]byte(raw), rule.OptionsBase{ConfigDirectory: "/repo", ProjectRoot: "/elsewhere"})
		if err != nil {
			t.Fatalf("Base's own block was refused: %v", err)
		}
		if root := decoded.(DependenciesOptions).Root; root != "/repo" {
			t.Fatalf("expected the root at the config's directory, got %q", root)
		}
	}
	decoded, err := decodeDependenciesOptions([]byte(projectModuleOptions), rule.OptionsBase{ProjectRoot: "/project"})
	if err != nil {
		t.Fatal(err)
	}
	if root := decoded.(DependenciesOptions).Root; root != "/project" {
		t.Fatalf("expected the project root when the config's directory is unknown, got %q", root)
	}

}
