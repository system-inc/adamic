# Markdown parser and layout: embedding off

The contract is Markdown parsing and layout with `embeddedLanguageFormatting: off`
on both oracles. The port is held to Go cohere. Embedded JSON, YAML, TOML and
other languages compose later from their own slices. Front matter and fenced
contents are raw under this contract. The prior inline printers are available as
a dependency. The complete native parser and block formatter are still unfinished.

## Native code block layout boundary

`codeblocks.ts` prints fenced and indented values with embedding off and computes
canonical fences without regular expressions. Go supplies AST fields and explicit
text display widths; parsing and preprocessing remain unported. Embedded JSON,
YAML, TOML and other languages remain raw, as required by this unit's contract.

## Native table layout boundary

`tables.ts` builds preserved-prose table rows and renders its child documents in
native code. Cell and full-row display widths are explicit inputs from Go's
Unicode width service. The compact `proseWrap: never` alternative is not ported.
The shared test now checks the original fork's complete Markdown parser/layout
with embedding off for all accumulated source cases, as well as doc printers.
Parsing/preprocessing and other unported children remain supplied by Go fixtures.

## Native quote layout boundary

`quotes.ts` constructs quotes and their child spacing from explicit AST facts.
Its native children include lists and eligible words. The fixture bridge supplies
Go-parsed/preprocessed trees, unported child docs and display widths. Tests cover
all 953 original contexts plus list and quote cases. This is a native layout
component, not yet a native Markdown-file parser or formatter. Parser state-frame
work and Unicode widths remain as described below. No new compiler gap was found.

## The default full-document oracles disagree

Go cohere's real `native.Formatter` invokes the Markdown parser/printer and its
native embedded-language dispatch. The JavaScript library it follows is its
vendored Prettier fork, based on version 3.9.6, not unmodified npm Prettier.
`testdata/library.mjs` calls those exact standalone/plugin bundles on Node.
The audit also measured stock npm Prettier 3.9.6 separately.

On the initial census of 873 physical repository/submodule Markdown files and
75 generated block cases, Go and the pinned fork matched 941/948 documents with
embedding enabled. Seven real files disagreed: one JSONC fence and six Flow
fixtures. Stock npm Prettier matched 788/948. With embedding disabled, both
JavaScript libraries matched Go on all 948. The test repeats the census and
checks a named list of observed default-mode gaps; closing or adding a gap fails
that assertion so the report cannot silently become stale.

`gaps/embedded_jsonc.md` is a minimal proving document. Go preserves the contents
because `native/json.go` registers only `.json`, while embedded `jsonc` maps to
`embedded.jsonc`. The fork formats it and inserts a trailing comma. One output
cannot equal both. `gaps/embedded_flow.md` proves the Flow embedding difference:
the native JavaScript path does not format that Flow program as the Babel/Flow
printer in the fork does. These dependencies are outside this unit's territory.
These auto-mode differences are historical witnesses, outside the agreed
embedding-off contract. Front matter is deliberately raw with embedding off.

The audit checks both `auto` and `off` without filtering out embedded files.
The off baseline calls Go's actual `markdown.Format` with a nil embedding
callback, and normalizes/restores BOM/line endings exactly at the core boundary.
The auto baseline calls its actual `native.Formatter`. Neither adapter duplicates
Markdown parsing or printing. The fixed options on both sides are tab width 4,
print width 120, single quotes, spaces, semicolons and preserved prose, with the
remaining Prettier defaults. This is an explicit common option set, not each
nested repository's configuration or ignore rules. Every physical Markdown file
is included, even files formatting enumeration would skip.

## The literal parser representation also needs a redesign

`micromark/types.go` represents states as `func(code Code) State`, with nil meaning
finished. The document/container/flow machines are mutually referring closures
with captured mutable frames. These are additional prerequisites for a direct
Adamic port, not evidence that Markdown cannot ever be implemented in Adamic:

- `gaps/1_recursive_state.ts` prints `35` twice on Node. Stage 0 returns exactly
  `lower.NotYet("a function inside a function (a closure)")` for the nested
  function declaration corresponding to a state with its captured frame.
- `gaps/2_state_arrow_cycle.ts` prints `35` twice on Node. Rewriting the nested
  declaration as a mutable arrow slot is refused by Adamic 0.1 as a closure/cell
  reference-counting cycle. This is a `Refused`, not a `NotYet`.
- `gaps/3_recursive_callable.ts` is closed: a noncapturing, top-level recursive
  callable state is supported. Source Node, native under ASan/UBSan, the
  JavaScript backend and LeakSanitizer agree on `1` then `0`. The recursive
  callable type itself is therefore not the blocker.

A full port can instead use explicit state/frame IDs and owned arenas, with
carefully bounded weak back references. That requires porting and checking the
micromark engine, every construct/extension and mdast compilation, followed by
preprocessing, AstPath and document layout. This branch does not pretend that a
line-oriented approximation is the same parser. No compiler files were modified.

## Reproduction and scope of the checks

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_MARKDOWNBLOCKS_FORK=/tmp/adamic-markdown-blocks-fork \
  ADAMIC_MARKDOWNBLOCKS_KEEP=/tmp/markdown-blocks-audit \
  go test -count=1 -v -timeout=30m ./stage1/cohere/markdownblocks > /tmp/markdown-blocks-test.log 2>&1
```

Install the original fork in that scratch directory by copying the exact
`cohere/internal/format/prettier/bundles/` tree from the pinned submodule. The
adapter checks `prettier.version === '3.9.6'`. The unit report records its pin and
bundle digests. This is an installation of the vendored original, not a guessed
patch to npm Prettier. The stock comparison uses the separately installed exact
npm version in `/tmp/adamic-markdown-prettier/node_modules/prettier`.

The census includes `.md`, `.markdown`, `.mdown` and `.mkd`, skipping only `.git`.
Generated cases cover heading forms, all list marker types and five nesting
depths, fenced/indented code, block quotes, tables, HTML, thematic breaks, front
matter and typed code fences. Three Go-printer mutants alter an ATX heading
prefix, unordered list marker and fence length through overlays. Each must build,
exit zero, produce no stderr or formatting errors, and fail the off baseline's
byte comparison. These prove the preflight comparison catches real output
changes; they are Go mutants, not mutants of an unbuilt Adamic block printer.

The list printer and the document operations Markdown uses are now ported.
The complete native Markdown parser, preprocessing, remaining block printers
and Unicode display-width service remain unfinished. The block unit remains incomplete. The prior inline implementation is
composed only as a branch dependency, not yet wired into a block formatter.
The embedding-off contract avoids those embedding dependencies. The remaining
work is the native parser/frame and document printer, not a change to embedded
language formatting.

## Native front-matter parser stage

`frontmatter.ts` ports the complete `mdast.ParseFrontMatter` stage. It recognizes
both YAML and TOML delimiters, explicit languages and YAML's `...` terminator,
retains the exact raw text, and blanks the prefix one space per UTF-16 unit while
preserving LF. The upstream closing-delimiter suffix check accepts all suffixes;
that behavior is retained. ECMAScript whitespace is scanned explicitly: U+FEFF
is trimmed and U+0085 is retained. This stage receives the parser's text, before
any micromark events; it does not normalize line endings or remove BOM itself.

`testdata/frontmatter_probe.ts` exercises this stage, not full Markdown layout.
Its test walks all 876 physical Markdown files, the 77 existing generated block
cases and 3,116 new front-matter cases. Go's actual `ParseFrontMatter`, the pinned
fork's real Markdown parser, source Node, Adamic's JavaScript backend and native
agree on all 4,069 observations. Native runs under ASan/UBSan and LeakSanitizer.
The fork parser's first `frontMatter` node supplies the original library fields;
its blanked prefix is checked with the library's specified non-Unicode JavaScript
replacement. No original parsing is delegated to an oracle in production code.

Three native output-only mutants change YAML fallback selection, add one blank
space to each prefix line, and remove BOM from the language whitespace class.
Each lowers, builds and exits zero with no stderr; only its bytes fail the oracle
comparison. These are native front-matter mutants, not full block-printer mutants.
The inherited Go heading/list/fence mutants still test the whole-document audit.
The prior inline printers are not wired into a block formatter yet.

## Native list layout component

`lists.ts` ports the list and list-item printers, ordered-marker scanning,
Git-friendly numbering and list alignment. It covers all marker families,
999,999,999 capping, alternating sibling markers, task boxes, tight/loose spacing,
blank lines before nested lists, required indentation before code, and the HTML
column exception. `document.ts` and `commandStack.ts` port the document operations
used by Markdown with owned arenas and numeric IDs for children and indentation
roots. Captured mutually referring state closures are not used.

The test collects real Go parser/preprocessor trees and child documents for
blocks not ported yet. Labels are transparent to Go's actual formatter; the
collector checks that fact on every input. In the native fixture, every list
label is replaced by native list construction, and every eligible word label by
the prior native word printer. The document renderer lays out the whole file.
Fixtures are serialized before Go propagates breaks, so native break propagation
is exercised. All text display widths are explicit inputs to this component;
Unicode width is still a caller service, not a newly ported native implementation.
The native component never invokes Go or Node, but its fixture driver is not a
Markdown-file formatter. Source parsing, non-list child printing, path/context
facts and inherited alignment still come from Go in the test harness. No native
CommonMark list recognition/container machine is claimed by this layout slice.

The current corpus is all 953 existing document cases plus 1,218 generated list
cases, 2,171 whole-document contexts. Generated inputs cross ten markers, six
space/tab prefixes, four task forms and five body forms, plus numbering, cap,
blank-line, nested code/quote/table/HTML, definitions and ignore edges. Native,
source Node, the JavaScript backend and the original fork's document printer
agree with actual Go Markdown formatting. Native ASan/UBSan and leaks pass.
Three native mutants change the unordered marker, checked task box and ordered
cap; each compiles, exits zero without stderr and fails only the byte comparison.
This proves list layout on Go trees, not completion of native Markdown parsing.

## Seven upstream embedding witnesses, with complete outputs

These are the seven existing repository files from the original census. Both
outputs use embedding **auto**, tab width 4, print width 120, preserved prose,
spaces, semicolons and single quotes. The Go pin is
`715ba94f3608a6500086b1076ce5cb7e51b836db`; the other output comes from its exact
vendored Prettier 3.9.6 fork. With embedding **off**, each pair agrees. The JSON
string literals below preserve every byte, including the final newline. Input
SHA-256 identifies the witness independently of later edits.

### 1. `cohere/TypeScript/packages/vscode-typescript/README.md`

Input SHA-256: `fdb86d928c1f51034f590027c1e158aaf092ddd0b587fac4f4f80ef46d2c8822`.

Go cohere, complete output as a JSON string:

````text
"# TypeScript 7\n\nThis extension provides the native implementation of the TypeScript language service. It provides features like go-to-definition, completions, errors and diagnostics, quick info/tooltip hovers, and more.\n\n## Usage\n\n1. Install the extension from the marketplace.\n2. Open a TypeScript or JavaScript file (`.ts`) in your editor.\n3. Activate the extension with the command `TypeScript: Enable TypeScript 7`, or update your settings below:\n\n## Configuration\n\nYou can enable this extension by modifying the following settings:\n\n```jsonc\n{\n    // UI Setting:\n    // TypeScript 7 > Experimental: Use Tsgo\n    \"js/ts.experimental.useTsgo\": true,\n\n    // Optional: use a local TypeScript package directory.\n    \"js/ts.tsdk.path\": \"./node_modules/typescript\"\n}\n```\n\n## Feedback\n\nIf you encounter any issues or have suggestions for improvement, please open an issue on the [GitHub repository](https://github.com/microsoft/TypeScript/tsc).\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"# TypeScript 7\n\nThis extension provides the native implementation of the TypeScript language service. It provides features like go-to-definition, completions, errors and diagnostics, quick info/tooltip hovers, and more.\n\n## Usage\n\n1. Install the extension from the marketplace.\n2. Open a TypeScript or JavaScript file (`.ts`) in your editor.\n3. Activate the extension with the command `TypeScript: Enable TypeScript 7`, or update your settings below:\n\n## Configuration\n\nYou can enable this extension by modifying the following settings:\n\n```jsonc\n{\n    // UI Setting:\n    // TypeScript 7 > Experimental: Use Tsgo\n    \"js/ts.experimental.useTsgo\": true,\n\n    // Optional: use a local TypeScript package directory.\n    \"js/ts.tsdk.path\": \"./node_modules/typescript\",\n}\n```\n\n## Feedback\n\nIf you encounter any issues or have suggestions for improvement, please open an issue on the [GitHub repository](https://github.com/microsoft/TypeScript/tsc).\n"
````

### 2. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_fn_type_mismatch_2.expect.md`

Input SHA-256: `8686b393bc7969930ea787b1a735c652ce2b029a544cd44cab152282614d0cf5`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n * @flow strict-local\n * @format\n */\n\n'use strict';\n\nimport type {SimpleTooltipMessageTypesType} from 'SimpleTooltipMessageTypes';\nimport type {Alignment, Position, Width} from 'AmbientTooltip.react';\n\nimport * as SimpleTooltipMessage from 'SimpleTooltipMessage';\nimport AmbientTooltip from 'AmbientTooltip.react';\n\nimport * as React from 'react';\n\nexport type NUXProps = {\n  alignment?: Alignment,\n  children: React.Node,\n  customWidth?: number,\n  disabled?: boolean,\n  hideOnXout?: boolean,\n  position: Position,\n  showOnce?: boolean,\n  type: SimpleTooltipMessageTypesType,\n  width?: Width,\n};\n\ntype CurrentState = {\n  showNux: boolean,\n};\n\ntype ExposedProps<Props extends {...}> = {\n  ...$Exact<Props>,\n  nuxProps: NUXProps,\n};\n\nexport default function WidgetWithTooltip<\n  Props extends {...},\n  WidgetWithTooltipComponent extends React.ComponentType<Props>,\n>(\n  WrappedComponent: WidgetWithTooltipComponent\n): Class<\n  React.Component<\n    ExposedProps<React.ElementConfig<WidgetWithTooltipComponent>>,\n    CurrentState,\n  >,\n> {\n  class WithNux extends React.PureComponent<ExposedProps<Props>, CurrentState> {\n    state: CurrentState = {\n      showNux:\n        this.props.nuxProps.disabled !== true &&\n        !SimpleTooltipMessage.hasUserSeenMessage_LEGACY(\n          this.props.nuxProps.type\n        ),\n    };\n\n    wrappedRef: {\n      current: HTMLSpanElement | null,\n      ...\n    } = React.createRef();\n\n    componentDidMount(): void {\n      if (this.props.nuxProps.showOnce === true && this.state.showNux) {\n        SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n      }\n    }\n\n    #onNuxClose = (): void => {\n      if (this.props.nuxProps.hideOnXout === true && this.state.showNux) {\n        SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n      }\n      this.setState({\n        showNux: false,\n      });\n    };\n\n    #getRef = (): null | HTMLSpanElement => this.wrappedRef.current;\n\n    render(): React.MixedElement {\n      const {nuxProps, ...passProps} = this.props;\n      return (\n        <>\n          <span className=\"uiContextualLayerParent\" ref={this.wrappedRef}>\n            <WrappedComponent {...passProps} />\n          </span>\n          {this.state.showNux ? (\n            <AmbientTooltip\n              alignment={nuxProps.alignment}\n              children={nuxProps.children}\n              contextRef={this.#getRef}\n              customwidth={nuxProps.customWidth}\n              onCloseButtonClick={this.#onNuxClose}\n              position={nuxProps.position}\n              shown={this.state.showNux}\n              width={nuxProps.width}\n            />\n          ) : null}\n        </>\n      );\n    }\n  }\n  return WithNux;\n}\n\n```\n\n## Error\n\n```\nUnexpected token (32:33)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n * @flow strict-local\n * @format\n */\n\n'use strict';\n\nimport type { SimpleTooltipMessageTypesType } from 'SimpleTooltipMessageTypes';\nimport type { Alignment, Position, Width } from 'AmbientTooltip.react';\n\nimport * as SimpleTooltipMessage from 'SimpleTooltipMessage';\nimport AmbientTooltip from 'AmbientTooltip.react';\n\nimport * as React from 'react';\n\nexport type NUXProps = {\n    alignment?: Alignment,\n    children: React.Node,\n    customWidth?: number,\n    disabled?: boolean,\n    hideOnXout?: boolean,\n    position: Position,\n    showOnce?: boolean,\n    type: SimpleTooltipMessageTypesType,\n    width?: Width,\n};\n\ntype CurrentState = {\n    showNux: boolean,\n};\n\ntype ExposedProps<Props: { ... }> = {\n    ...$Exact<Props>,\n    nuxProps: NUXProps,\n};\n\nexport default function WidgetWithTooltip<Props: { ... }, WidgetWithTooltipComponent: React.ComponentType<Props>>(\n    WrappedComponent: WidgetWithTooltipComponent,\n): Class<React.Component<ExposedProps<React.ElementConfig<WidgetWithTooltipComponent>>, CurrentState>> {\n    class WithNux extends React.PureComponent<ExposedProps<Props>, CurrentState> {\n        state: CurrentState = {\n            showNux:\n                this.props.nuxProps.disabled !== true &&\n                !SimpleTooltipMessage.hasUserSeenMessage_LEGACY(this.props.nuxProps.type),\n        };\n\n        wrappedRef: {\n            current: HTMLSpanElement | null,\n            ...\n        } = React.createRef();\n\n        componentDidMount(): void {\n            if(this.props.nuxProps.showOnce === true && this.state.showNux) {\n                SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n            }\n        }\n\n        #onNuxClose = (): void => {\n            if(this.props.nuxProps.hideOnXout === true && this.state.showNux) {\n                SimpleTooltipMessage.markMessageSeenByUser(this.props.nuxProps.type);\n            }\n            this.setState({\n                showNux: false,\n            });\n        };\n\n        #getRef = (): null | HTMLSpanElement => this.wrappedRef.current;\n\n        render(): React.MixedElement {\n            const { nuxProps, ...passProps } = this.props;\n            return (\n                <>\n                    <span className=\"uiContextualLayerParent\" ref={this.wrappedRef}>\n                        <WrappedComponent {...passProps} />\n                    </span>\n                    {this.state.showNux ? (\n                        <AmbientTooltip\n                            alignment={nuxProps.alignment}\n                            children={nuxProps.children}\n                            contextRef={this.#getRef}\n                            customwidth={nuxProps.customWidth}\n                            onCloseButtonClick={this.#onNuxClose}\n                            position={nuxProps.position}\n                            shown={this.state.showNux}\n                            width={nuxProps.width}\n                        />\n                    ) : null}\n                </>\n            );\n        }\n    }\n    return WithNux;\n}\n```\n\n## Error\n\n```\nUnexpected token (32:33)\n```\n"
````

### 3. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_loc_diff_1.expect.md`

Input SHA-256: `f40f1db9e8a46fc284bee4df1ac6bccb6cfefdec0b9df25159bb7c8211514fb6`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n *  * @flow strict\n * @format\n */\n\n/**\n * creates a cache for Component props so we prevent rendering a component\n * sequentially if the props didn't change. Useful to wrap FluxContainer\n * with pure calculateState and getStores functions.\n */\n\n'use strict';\n\nimport * as React from 'react';\nimport {PureComponent} from 'react';\n\nexport default function createPureComponent<\n  DefaultProps extends {...} | void,\n  Props extends {...},\n>(\n  Component: React.ComponentType<Props> & {\n    defaultProps?: DefaultProps,\n    displayName?: string,\n  }\n): React.ComponentType<Props> {\n  class PureComponentCache extends PureComponent<Props, void> {\n    static defaultProps: DefaultProps;\n\n    render(): React.MixedElement {\n      return <Component {...this.props} />;\n    }\n  }\n\n  if (Component.defaultProps) {\n    PureComponentCache.defaultProps = Component.defaultProps;\n  }\n  PureComponentCache.displayName = `PureComponentCache(${\n    /* $FlowFixMe[incompatible-type] (>=0.66.0 site=www) This comment\n     * suppresses an error found when Flow v0.66 was deployed. To see the\n     * error delete this comment and run Flow. */\n    Component.displayName\n  })`;\n  // $FlowFixMe[incompatible-type]\n  return PureComponentCache;\n}\n\n```\n\n## Error\n\n```\nUnexpected token (18:24)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n/**\n *  * @flow strict\n * @format\n */\n\n/**\n * creates a cache for Component props so we prevent rendering a component\n * sequentially if the props didn't change. Useful to wrap FluxContainer\n * with pure calculateState and getStores functions.\n */\n\n'use strict';\n\nimport * as React from 'react';\nimport { PureComponent } from 'react';\n\nexport default function createPureComponent<DefaultProps: { ... } | void, Props: { ... }>(\n    Component: React.ComponentType<Props> & {\n        defaultProps?: DefaultProps,\n        displayName?: string,\n    },\n): React.ComponentType<Props> {\n    class PureComponentCache extends PureComponent<Props, void> {\n        static defaultProps: DefaultProps;\n\n        render(): React.MixedElement {\n            return <Component {...this.props} />;\n        }\n    }\n\n    if(Component.defaultProps) {\n        PureComponentCache.defaultProps = Component.defaultProps;\n    }\n    PureComponentCache.displayName = `PureComponentCache(${\n        /* $FlowFixMe[incompatible-type] (>=0.66.0 site=www) This comment\n         * suppresses an error found when Flow v0.66 was deployed. To see the\n         * error delete this comment and run Flow. */\n        Component.displayName\n    })`;\n    // $FlowFixMe[incompatible-type]\n    return PureComponentCache;\n}\n```\n\n## Error\n\n```\nUnexpected token (18:24)\n```\n"
````

### 4. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-pattern3_type_to_poly.expect.md`

Input SHA-256: `09e9ebd9520468174c33b4479402d511f08d097ebe2611efd5db961961b4da05`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Pattern 3: shapeId null→generated + return Type→Poly\n// Generic function with Object.entries().reduce()\n// Divergence: TS has shapeId:null, return:Type(32); Rust has shapeId:\"<generated_1>\", return:Poly\n\n/**\n * @flow strict\n */\nexport default function flipAndAggregateObject<TValue extends string>(obj: {\n  +[key: TKey]: TValue,\n  ...\n}): {} {\n  return Object.entries(obj).reduce((acc, [currKey, currVal]) => {\n    return {};\n  }, {});\n}\n\n```\n\n## Error\n\n```\nUnexpected token (9:2)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Pattern 3: shapeId null→generated + return Type→Poly\n// Generic function with Object.entries().reduce()\n// Divergence: TS has shapeId:null, return:Type(32); Rust has shapeId:\"<generated_1>\", return:Poly\n\n/**\n * @flow strict\n */\nexport default function flipAndAggregateObject<TValue: string>(obj: { +[key: TKey]: TValue, ... }): {} {\n    return Object.entries(obj).reduce((acc, [currKey, currVal]) => {\n        return {};\n    }, {});\n}\n```\n\n## Error\n\n```\nUnexpected token (9:2)\n```\n"
````

### 5. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_identifier_diff.expect.md`

Input SHA-256: `53c0cd7ff1745679ace6a7ee83e9470eea3f6f2b9b82cccfaeeccbec8f4d8be5`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: IDENTIFIER_DIFF (11 files)\n// Extra/different context identifiers — class components with this/setState\n/**\n * @flow strict-local\n */\nexport default function withRemountOnChange<OuterProps extends {}>(\n  shouldRemount: () => boolean\n): () => React.ComponentType<OuterProps> {\n  return function withRemountOnChangeInner(WrappedComponent) {\n    return class Wrapper extends React.Component<OuterProps, WrapperState> {\n      static displayName: ?string = `withRemountOnChange(${getDisplayName()})`;\n      state: WrapperState = {};\n      componentDidUpdate(prevProps: OuterProps, prevState: WrapperState) {\n        if (shouldRemount()) {\n          this.setState(({keyId}) => {});\n        }\n      }\n      render(): React.MixedElement {}\n    };\n  };\n}\n\n```\n\n## Error\n\n```\nUnexpected token (11:26)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: IDENTIFIER_DIFF (11 files)\n// Extra/different context identifiers — class components with this/setState\n/**\n * @flow strict-local\n */\nexport default function withRemountOnChange<OuterProps: {}>(\n    shouldRemount: () => boolean,\n): () => React.ComponentType<OuterProps> {\n    return function withRemountOnChangeInner(WrappedComponent) {\n        return class Wrapper extends React.Component<OuterProps, WrapperState> {\n            static displayName: ?string = `withRemountOnChange(${getDisplayName()})`;\n            state: WrapperState = {};\n            componentDidUpdate(prevProps: OuterProps, prevState: WrapperState) {\n                if(shouldRemount()) {\n                    this.setState(({ keyId }) => {});\n                }\n            }\n            render(): React.MixedElement {}\n        };\n    };\n}\n```\n\n## Error\n\n```\nUnexpected token (11:26)\n```\n"
````

### 6. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_severity_diff.expect.md`

Input SHA-256: `4e4b4b0a786f3ae7de1266c8e4786b107cb0aa2edc753debe2f755aa445f764b`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: SEVERITY_DIFF (1 file from OTHER)\n// delete on optional chain — TS: severity Error/Syntax, Rust: severity Hint/Todo\n/**\n * @flow strict-local\n */\nimport {} from 'PreloadingTTL';\nconst preloadedRequests: Map<> = new Map();\nexport function execute(): Promise<{error?: APIErrorEventArgs['error'], ...}> {\n  if (request.params != null && !(request.params instanceof FormData)) {\n    delete request.params?.__entryPointPreloaded;\n  }\n  if (!consumers) {\n    if (APIRequestMatchingUtils.areRequestsEquivalent()) {\n    }\n  }\n}\n\n```\n\n## Error\n\n```\nType argument list cannot be empty. (7:28)\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// Round 2 HIR: SEVERITY_DIFF (1 file from OTHER)\n// delete on optional chain — TS: severity Error/Syntax, Rust: severity Hint/Todo\n/**\n * @flow strict-local\n */\nimport {} from 'PreloadingTTL';\nconst preloadedRequests: Map<> = new Map();\nexport function execute(): Promise<{ error?: APIErrorEventArgs['error'], ... }> {\n    if(request.params != null && !(request.params instanceof FormData)) {\n        delete request.params?.__entryPointPreloaded;\n    }\n    if(!consumers) {\n        if(APIRequestMatchingUtils.areRequestsEquivalent()) {\n        }\n    }\n}\n```\n\n## Error\n\n```\nType argument list cannot be empty. (7:28)\n```\n"
````

### 7. `cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-update-expression-context-variable-via-type-annotation.expect.md`

Input SHA-256: `cd4a073e669a886fc32f65b5ec66b11ae6fb3f2972e9074936ca0c9bc222ccea`.

Go cohere, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// @flow @compilationMode(infer)\nfunction Component(props: {data: Array<[string, mixed]>}) {\n  let id = 0;\n  for (const [key, value] of props.data) {\n    const item = {\n      key,\n      id: '' + id++,\n    };\n  }\n  const getIndex = ((): ((id: string) => number) => {\n    return (id: string): number => 0;\n  })();\n  return <div />;\n}\n\n```\n\n## Error\n\n```\nFound 1 error:\n\nTodo: (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n\n   5 |     const item = {\n   6 |       key,\n>  7 |       id: '' + id++,\n     |                ^^^^ (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n   8 |     };\n   9 |   }\n  10 |   const getIndex = ((): ((id: string) => number) => {\n```\n"
````

Cohere's Prettier fork, complete output as a JSON string:

````text
"## Input\n\n```javascript\n// @flow @compilationMode(infer)\nfunction Component(props: { data: Array<[string, mixed]> }) {\n    let id = 0;\n    for(const [key, value] of props.data) {\n        const item = {\n            key,\n            id: '' + id++,\n        };\n    }\n    const getIndex = ((): ((id: string) => number) => {\n        return (id: string): number => 0;\n    })();\n    return <div />;\n}\n```\n\n## Error\n\n```\nFound 1 error:\n\nTodo: (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n\n   5 |     const item = {\n   6 |       key,\n>  7 |       id: '' + id++,\n     |                ^^^^ (BuildHIR::lowerExpression) Handle UpdateExpression to variables captured within lambdas.\n   8 |     };\n   9 |   }\n  10 |   const getIndex = ((): ((id: string) => number) => {\n```\n"
````
