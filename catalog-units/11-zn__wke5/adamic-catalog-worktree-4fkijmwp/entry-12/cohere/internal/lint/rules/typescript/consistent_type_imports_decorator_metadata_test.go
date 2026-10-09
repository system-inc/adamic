package typescript

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * Decorator metadata turns a type position into a runtime reference.
 *
 * Under `emitDecoratorMetadata`, a decorated declaration's types are serialized into
 * `design:paramtypes`, `design:type` and `design:returntype`, and a type naming a class is emitted as
 * a reference to that class. Rewriting such an import to `import type` erases the binding, the
 * metadata becomes `Object`, and nothing fails until a framework reads it at run time.
 *
 * The shapes are api-phi-health's `source/modules/forms/internal/entities/FormComponent.ts`, whose
 * `@GraphQlObjectType` classes take a `FormComponentCreateInput` (a class) in their constructors and
 * whose Base GraphQL layer reads the metadata in `FindType.ts` and `GraphQlField.ts`. That tree
 * reported 1,295 findings here against ESLint's zero.
 *
 * The compiler keeps an import for metadata only when it names a value and is not a const enum, so
 * an interface in the same position still reports: under `isolatedModules` tsc itself demands
 * `import type` for it (TS1272). The last case is the control proving the option is what silences
 * the class.
 */

// decoratorMetadataConfig is the harness's default plus the three options api-phi-health's Base
// configuration sets for decorators.
const decoratorMetadataConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"module": "esnext",
		"moduleDetection": "force",
		"types": [],
		"experimentalDecorators": true,
		"emitDecoratorMetadata": %s,
		"isolatedModules": true
	},
	"include": ["**/*.ts", "**/*.tsx"]
}`

const decoratorMetadataTypes = `export class FormComponentCreateInput { kind = ''; }
export class Ordinary { kind = ''; }
export interface FormComponentShape { kind: string; }
export const enum FormComponentKind { Text, Number }
`

const decoratorMetadataDecorators = `export function GraphQlObjectType(): ClassDecorator { return () => {}; }
export function GraphQlField(): PropertyDecorator { return () => {}; }
export function GraphQlMethod(): MethodDecorator { return () => {}; }
export function GraphQlArgument(): ParameterDecorator { return () => {}; }
`

func runDecoratorMetadata(t *testing.T, subject string, emitDecoratorMetadata bool) rule_testing.Result {
	t.Helper()
	enabled := "false"
	if emitDecoratorMetadata {
		enabled = "true"
	}
	return rule_testing.RunTypedFilesWithSetupAndOptions(t, ConsistentTypeImports, map[string]string{
		"FormServiceTypes.ts": decoratorMetadataTypes,
		"decorators.ts":       decoratorMetadataDecorators,
		"FormComponent.ts":    subject,
	}, "FormComponent.ts", DefaultConsistentTypeImportsOptions(), func(directory string) {
		config := []byte(fmt.Sprintf(decoratorMetadataConfig, enabled))
		if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), config, 0o644); err != nil {
			t.Fatalf("writing the decorator metadata tsconfig: %v", err)
		}
	})
}

func TestConsistentTypeImportsKeepsDecoratorMetadataReferences(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		ids    []string
	}{
		{
			name: "FormComponent.ts: a decorated class's constructor takes an imported class",
			source: `import { GraphQlObjectType } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
@GraphQlObjectType()
export class FormComponentData {
    constructor(input: FormComponentCreateInput, position: number) { void input; void position; }
}
`,
		},
		{
			name: "the same file still reports an ordinary type-only import",
			source: `import { GraphQlObjectType } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
import { Ordinary } from './FormServiceTypes';
@GraphQlObjectType()
export class FormComponentData {
    constructor(input: FormComponentCreateInput) { void input; }
}
export let other: Ordinary;
`,
			ids: []string{"typeOverValue"},
		},
		{
			name: "a decorated property's type",
			source: `import { GraphQlField } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
export class FormComponentData {
    @GraphQlField()
    input!: FormComponentCreateInput;
}
`,
		},
		{
			name: "a decorated method's parameter and return types",
			source: `import { GraphQlMethod } from './decorators';
import { FormComponentCreateInput, Ordinary } from './FormServiceTypes';
declare const stored: unknown;
export class FormResolver {
    @GraphQlMethod()
    create(input: FormComponentCreateInput): Ordinary { void input; return stored as Ordinary; }
}
`,
		},
		{
			name: "a decorated getter's return type",
			source: `import { GraphQlMethod } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
declare const stored: unknown;
export class FormComponentData {
    @GraphQlMethod()
    get input(): FormComponentCreateInput { return stored as FormComponentCreateInput; }
}
`,
		},
		{
			// The checker falls back to the opposite accessor's annotation when the decorated one has
			// none, so an unannotated setter is serialized from its getter.
			name: "a decorated setter without an annotation takes its getter's",
			source: `import { GraphQlMethod } from './decorators';
import { Ordinary } from './FormServiceTypes';
declare const stored: unknown;
export class FormComponentData {
    get value(): Ordinary { return stored as Ordinary; }
    @GraphQlMethod()
    set value(next) { void next; }
}
`,
		},
		{
			name: "a decorated parameter serializes its whole signature",
			source: `import { GraphQlArgument } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
export class FormResolver {
    create(@GraphQlArgument() position: number, input: FormComponentCreateInput) { void position; void input; }
}
`,
		},
		{
			name: "an interface in a decorated signature still reports, as tsc's TS1272 demands",
			source: `import { GraphQlObjectType } from './decorators';
import { FormComponentShape } from './FormServiceTypes';
@GraphQlObjectType()
export class FormComponentData {
    constructor(input: FormComponentShape) { void input; }
}
`,
			ids: []string{"typeOverValue"},
		},
		{
			name: "a class joined with null under strict null checks serializes as Object, so it reports",
			source: `import { GraphQlField } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
export class FormComponentData {
    @GraphQlField()
    input!: FormComponentCreateInput | null;
}
`,
			ids: []string{"typeOverValue"},
		},
		{
			name: "a const enum is inlined, so it reports",
			source: `import { GraphQlField } from './decorators';
import { FormComponentKind } from './FormServiceTypes';
export class FormComponentData {
    @GraphQlField()
    kind!: FormComponentKind;
}
`,
			ids: []string{"typeOverValue"},
		},
		{
			name: "an undecorated class keeps no metadata, so it reports",
			source: `import { FormComponentCreateInput } from './FormServiceTypes';
export class FormComponentData {
    constructor(input: FormComponentCreateInput) { void input; }
}
`,
			ids: []string{"typeOverValue"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runDecoratorMetadata(t, testCase.source, true)
			if len(testCase.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}

	// The control. Without the option nothing reaches run time, so the class import that is silent
	// above is an ordinary type-only import and reports.
	t.Run("the same constructor without emitDecoratorMetadata reports", func(t *testing.T) {
		t.Parallel()
		result := runDecoratorMetadata(t, `import { GraphQlObjectType } from './decorators';
import { FormComponentCreateInput } from './FormServiceTypes';
@GraphQlObjectType()
export class FormComponentData {
    constructor(input: FormComponentCreateInput) { void input; }
}
`, false)
		rule_testing.ExpectFindings(t, result, "typeOverValue")
	})
}
