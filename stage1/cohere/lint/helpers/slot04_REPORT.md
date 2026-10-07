Built: JSX ElementParts, Tailwind settings key, compiled-reader cache (three .a helpers).
Implementation commit: 27ffb6a; claims were pushed before implementation.
Checks: helpers PASS 167.977s; uncached input oracle PASS 8.732s; go vet and gofmt clean; setup 165s, nproc 5.
Mutants: three compiling semantic mutants caught by Go comparisons on Node and sanitized native; two consumer-omission controls caught.
Not covered: full repository gate, production reader factory/regex engine, concurrent cache creation, direct compiler AST integration, malformed UTF-8 Go strings.

## Scope and readiness

48 dependency edges removed across 36 distinct consumers. Against the frozen original readiness ledger, 46 helper-ready rules become 47. Only `@next/next/no-img-element` loses its final helper blocker. With the already delivered comments bundle's 62-ready baseline, this single additional rule gives 63. These numbers describe helper prerequisites, not completed rule implementations. The original readiness.json remains unchanged; slot04_readiness.json records all residual dependencies.

### ecmascript/jsx.ElementParts

File: `jsx_element_parts.a`. 24 consumer dependencies removed. Final blockers removed: @next/next/no-img-element.

- `@next/next/google-font-display`
- `@next/next/google-font-preconnect`
- `@next/next/next-script-for-ga`
- `@next/next/no-before-interactive-script-outside-document`
- `@next/next/no-css-tags`
- `@next/next/no-head-element`
- `@next/next/no-html-link-for-pages`
- `@next/next/no-img-element`
- `@next/next/no-page-custom-font`
- `@next/next/no-styled-jsx-in-document`
- `@next/next/no-sync-scripts`
- `@next/next/no-unwanted-polyfillio`
- `react/forbid-dom-props`
- `react/jsx-key`
- `react/jsx-no-duplicate-props`
- `react/jsx-no-script-url`
- `react/jsx-no-target-blank`
- `react/jsx-props-no-spread-multi`
- `react/no-danger`
- `react/no-unknown-property`
- `react/void-dom-elements-no-children`
- `structure/boundary-no-project-theme-value`
- `structure/react-element-no-anchor`
- `structure/react-element-no-horizontal-rule`

### rules/tailwind.ClassLiteralSettings.key

File: `tailwind_settings_key.a`. 12 consumer dependencies removed. Final blockers removed: none.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-concatenated-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

### rules/tailwind.compiledClassLiteralReader

File: `tailwind_compiled_reader.a`. 12 consumer dependencies removed. Final blockers removed: none.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-concatenated-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

## Observed validation

Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Parser-derived node links and actual Go helper entry points supply expected output independently of Adamic. Node source and sanitized native both match:

- JSX: 1,083 captured inputs from all 24 consumers, 20,331 output lines; eight shape inputs, 130 lines.
- Settings key: five captured distinct configurations and 467 boundary configurations, including list order, duplicate/empty names, Unicode and embedded separators.
- Compiled reader: 323 recorded actual calls from all 12 Tailwind consumers and 364 control calls, including repeated keys and first-settings-wins collisions. Identity and factory invocation counts are checked.

Commands, with every test's output redirected to a file:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/04/helpers-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > stage1/cohere/lint/helpers/evidence/04/oracle.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/evidence/04/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot04_test.go stage1/cohere/lint/helpers/testdata/slot04 > stage1/cohere/lint/helpers/evidence/04/gofmt.log 2>&1
```

The full helper package passed in 167.977s. The input oracle's six fixtures passed in 8.732s, with six probe misses and zero cache hits. Vet and format output are empty and successful. Captured Next, React and structure fixture packages passed; the targeted Tailwind fixture selection passed. The entire repository test gate was not run.

## Mutants and controls

Each new semantic mutant compiled and ran successfully with no stderr, then disagreed with actual Go on both Node source and sanitized native:

| Mutation | Catch |
| --- | --- |
| Drop JSX self-closing case | Line 6: `parts -1 -1` instead of `parts 6 7`. |
| Omit first settings SOH separator | Line 1: bytes `1,` instead of `1,1,`. |
| Recreate reader on cache hit | Line 2: reader identity/count `2 2` instead of `1 1`. |

Deleting all captured rows for `@next/next/google-font-display` and for `better-tailwindcss/enforce-canonical-classes` separately makes the consumer-coverage checker fail. These are metadata omission controls, distinct from compiling semantic mutants.

The package gate also reran and caught the four inherited helper mutants: accept raw JSON newline (line 15157), wrong oneOf handling (15166), omit message interpolation (22166), and accept unknown strict option (15178). Ten inherited message-refusal cases also match Go, Node and sanitized native.

## Boundaries and workarounds

The cache is an explicit instance shared once per run, with an externally supplied factory. This matches Go's serial observable cache behavior; the production reader factory, regex compiler, and concurrent LoadOrStore creation races remain separate work. The arena is a stable-ID adapter to parser node fields, not an integration into Adamic's compiler AST. Settings keys preserve Go's separator collisions rather than inventing a new encoding. Invalid raw UTF-8 Go strings are outside the decoded configuration-string boundary.

Fixture capture normalizes Windows filenames before parsing. JSONL records split only at LF, because Unicode line separators can occur inside source strings. Tailwind's fixture tests had hardcoded developer checkout paths and initially failed here. Capture installs pinned tailwindcss@4.3.3 with scripts disabled into a temporary fixture and redirects those fixture roots through a Go overlay. Live-repository population tests are excluded by the documented capture selector. No cohere source edits are delivered. See slot04_README.md and capture.py for the reproducible selector and adapter details.

All helper branches were wildcard-fetched before selections. Claims use consumer counts from frozen readiness.json. Earlier concurrent claims for component-base, isSpace and ListenerKinds were yielded to earlier timestamped claims, as recorded in claims/04.md. Component-base's local passing experiment is excluded from delivery and totals; no isSpace or ListenerKinds implementation was written here. Final delivered reservations are JSX ElementParts (f48ee78), settings key (a38f2ed), and compiled reader (fd9dd3c). The final remote refresh found no competing compiled-reader claim.

## Setup

`bash cloud/setup.sh` succeeded; environment is /workspace/adamic-tools/env.sh. `nproc` returned 5.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (2s)
setup: build cache warm (165s)
setup: done in 165s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Logs in evidence/04 are durable evidence. Go 1.27.1, clang 20.1.8, Node 24.19.0. No pull request opened and no protected compiler files changed.
