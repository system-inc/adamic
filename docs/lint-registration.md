# Lint registration without shared lines

The shared scanner retains main's thirty syntax rules and discovers registered
rule modules beside them. The initial five registered ports own their messages,
options adapters, witnesses and mutants in `stage1/cohere/lint/rules/<slug>/`.

Add only your own rule directory. Do not append imports, dispatch conditions,
Go oracle selections, corpus filters, file-copy lists or mutant lists elsewhere.
The rule directory contains:

- `rule.a` (or existing `rule.ts`): an exported concrete listener class and named factory.
- `messages.a` (or existing `messages.ts`): exact descriptions and any message builders.
- `rule.json`: public name, interested node kinds, factory/class names and provenance.
- `oracle.go`: the unmodified upstream cohere rule and its typed options adapter.
- `testdata/*.{ts,tsx,js,jsx}.txt`: raw source witnesses, outside the module graph.
- `testdata/<name>.options.json` (optional, beside `<name>.ts.txt`): the rule's options for that witness,
  as upstream writes them. A rule that reports nothing by default, such as one that bans only the
  types it is configured with, witnesses a finding this way. The options reach the witness's own row
  (never its "all" row), so the Go adapter must decode them and the port reads the same settings.
- `mutant.json`: one source mutation that compiles, runs, and must disagree with Go.
  Fields are `name`, `from`, `to`, and optional `file` (default the discovered `rule.a` or `rule.ts`). The
  file must stay inside this rule directory. Anchors are scoped to that file,
  so another worker can use the same snippet without changing your test.

For example, the debugger descriptor is:

```json
{
  "name": "no-debugger",
  "order": 1,
  "kinds": ["DebuggerStatement"],
  "visit": "visit",
  "parent": true,
  "factory": "create",
  "class": "Rule",
  "oracle": "oracleNoDebugger",
  "upstreamPackage": "core",
  "upstreamTest": "TestNoDebugger"
}
```

`order` preserves the five migrated rules' prior dispatch and stable finding tie
order. **New rules omit it** and sort lexically by public name after migrated
rules. There is no shared ordinal allocator. Slugs contain lowercase letters,
digits and hyphens; scoped public names such as `@typescript-eslint/example`
stay in `name`. Upstream test names are prefixes because the tests have suffixes.
Only captured cases with the descriptor's exact rule name enter its corpus.

`create(context: RuleContext)` returns the exported concrete class. The class
implements `visit(index: number, parent: number)` when `parent: true`, or
`visit(index: number)` otherwise. Optional `prepare` and `finish` descriptor
fields name methods taking the source root index. File listeners subscribe to
`SourceFile`. Factories create one instance per source file; instances can own
rule-local state. Each file's full numeric ancestry is ready before `prepare`.
All prepare hooks run before preorder visitation; all finish hooks run afterward.
Hooks and visits run only for `all` or the exact selected name.

The shared context owns source, initialized parser/scanner, read-only parent
indexes, the finding destination, decoded JSON settings and main's existing equality/catch options.
It has no link back to Linter or RuleSet. Concrete named imports and typed fields
avoid interface-method casts and reference cycles. Keep new options decoding
inside the rule adapter. Manifest field 5 carries the complete captured options
as JSON, without a fixed per-rule capture struct. `context.settings.read` and
`list` expose scalar/string-array settings with lowercase keys. Empty/null input
uses defaults. The settings parser reads flat scalar/array values; structured values remain
JSON text for a rule's own decoder, and `settings.text` retains the complete raw
options. Rules can decode new shapes entirely within their own directory.
Option expressions are never executed. This is the
scanner branch's existing restricted settings parser, reused on main. Legacy
fields remain available, and JSON values override them when provided.

The adapter is built through a Go overlay inside cohere, with the generated
selection file and **all** discovered adapter files passed to `go build`. It
returns the upstream `rule.Rule` from a uniquely named function and has a second
function named `<oracle>Options(fields []string) any`. Preserve upstream defaults
and legacy manifest fields: equality mode/null are fields 2/3, empty-catch is field
4. JSON options are field 5; decode them to the rule's own upstream options
type in its adapter. For an object, `json.Unmarshal` retains the upstream
struct's decoded defaults; option validation remains the upstream rule's job.
Mark the adapter `//go:build lintoracle`: ordinary Adamic package discovery
must not build cohere's internal imports. The explicit-file overlay build includes
it. Cohere's rule bodies, messages, formatter and converging fixer stay independent
from the TS port and unchanged in the submodule.

From the repository root, before every manual build or Node run:

```sh
go run ./cmd/lint-registry
go run ./cmd/adamic build stage1/cohere/lint/main.ts
```

The generator validates the entire discovered set: strict JSON, unique public
names and adapter functions, complete directories, named factories/classes/hooks,
pinned AST kind names, required adapter exports, witnesses and mutant metadata.
It writes bucketed TS dispatch and Go selection into ignored `.generated/` files.
Generation uses sorted descriptors and node kinds, atomic replacement and unchanged
byte checks. A removed descriptor cannot leave a stale registration behind.
Do not commit these outputs. CI tests validate and regenerate before filtered tests
as well as before each port build and Node run.

Tests derive upstream package/test filters and copied module files by discovery.
Every owned witness must cause an upstream finding, and Go/Node/emitted-JavaScript/sanitized-native
outputs must match for selected and all-rule runs. Each discovered mutant is tested
against its own witnesses plus the inherited corner-case corpus. The existing
217 upstream cases, decoded defaults, exact descriptions, UTF-8 ranges, stable
findings, safe fixes, suggestions, fixed-source reparsing and overlap refusal are
preserved. This unit does not expand main's fix engine or parser recovery contract.

For a new rule, run the generator and package tests, review the upstream comparison
and mutant log, and commit only that directory. Two workers choosing distinct slugs
share no changed lines. Trial commits, merge output and tests are recorded in
[the evidence report](lint-registration-evidence/REPORT.md).

Oracle fixtures have their own companion migration on `codex/no-shared-lists`,
commit `7f958de`. Its source/options/counts sidecars and discovery command remove
fixture registry and table appends. This lint branch starts from main and does not
include that unit's changes; follow that unit's fixture instructions once merged.

## Adamic modules and shared certification

To use Adamic, rename your entry module to `rule.a`. Do not leave `rule.ts`
beside it: ambiguous entries are rejected. Rename other owned modules to `.a`
and update their explicit imports if desired. No descriptor changes are needed.
If `mutant.json` has an explicit `file`, update its extension; when omitted it
follows the entry module. Existing `.ts` entries and mutants still work.
Run `go run ./cmd/lint-registry` before a manual build. The default harness copies
both extensions and compares source Node, Adamic's emitted JavaScript on Node,
and ASan/UBSan native against unchanged Go cohere. Profiling uses the same copied
module graph and regenerates its registry.

Suggestions never become automatic fixes. Existing single-edit findings retain
their protocol. For complete suggestions, import `Suggestion` and `SuggestionEdit`
from `../../suggestions.a` and attach them to the `Finding` returned by
`context.report(...)`. Edits take UTF-16 source positions; the driver converts them
to Go's byte offsets. Preserve upstream suggestion order, ids, descriptions and
edit order. The driver serializes single edits as before, simple multi-edit or
out-of-finding suggestions as `suggestion-edits:<id>`, and multiple suggestions
or delimiter-bearing edits as separate `suggestion` and `suggestion-edit` rows.
The wave-05 `suggestion-edits:<id>` convention remains accepted without changes.
No shared serializer or profile edits are needed in a rule branch.

Raw witnesses keep their script extension, including `.tsx.txt` and `.jsx.txt`.
The harness preserves captured corpus filenames and Go script kinds as well.
This transports script kinds; it does not add JSX syntax to the shared stage 1 parser.
Unsupported syntax still fails the comparison instead of earning certification.

## Check one rule

From the repository root, run:

```sh
go run ./cmd/adamic-lint-check <slug>
```

The command validates every descriptor, then uses the shared registry generator
with only the selected registration. It checks owned witnesses, captured upstream
cases and inherited corner cases on Go cohere, source Node, emitted JavaScript and
ASan/UBSan native, including exact findings, repairs and suggestions. The owned
mutant must disagree with Go on all three port runtimes. A phase timing table is
printed even when certification fails.

Oracle binaries, captures, corpus observations and port artifacts are keyed by
inputs and toolchains under the user cache's `adamic/lint/` directory, published
atomically and checked for corruption. Selected native builds use the split
compiler with `Jobs = GOMAXPROCS`; correct and mutant builds run concurrently.
`ADAMIC_GATE_UNCACHED=1` bypasses observations, artifacts and split object reuse.
The ordinary all-rule suite keeps its full registry, cross-rule cases, mutants,
profiling checks and independent whole-file sanitized build:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 90m ./stage1/cohere/lint/...
```

All migrated rules now live in the registry, so the old hand-maintained baseline
selection and extra volume corpus are unnecessary. Their inherited sources and
options already live in the current shared corpus. Selected/full verdict parity
covers passing rules and a clean-running rule with a wrong message-ID literal.

## REPORT.md

Measured on the same cloud box at code tip `2df83490d475d7d09febb0115ac6b6a6c4d643ad`.
The final report commit changes documentation only. Fast cases are interleaved,
best of three; the full suite exceeds five minutes and was measured once.
Cold means an empty Adamic cache, with Go's build cache already populated.
Edits affect private snapshots of `rules/eqeqeq/rule.ts`, never the checkout:
one appended space, or the single literal `'unexpected'` to `'unexpectee'`.

Build flags for every row: `GOMAXPROCS=4`, `nproc=5`, cgroup `cpu.max=400000 100000`,
`go version go1.27.1 linux/amd64`, `clang version 20.1.8` (LLVM
`87f0227cb60147a26a1eeb4fb06e3b505e9c7261`), `node v24.19.0`.
Native uses `native.Flags(Options{Sanitize:true})`: C11, strict warnings,
`-ffp-contract=off -fno-optimize-sibling-calls -O1 -g
-fsanitize=address,undefined -fno-sanitize-recover=all`.
Fast builds use `Split:true, Jobs:4`; the uncached full suite uses whole-file C.

| Loop | Seconds | New objects | Cache mode | Load average before / after |
|---|---:|---:|---|---|
| Selected cold | 37.10 | 34 | Enabled, empty | 2.58 1.74 1.33 / 2.81 1.87 1.39 |
| Warm whitespace edit | 13.85 | 0 | Enabled, warm | 2.81 1.87 1.39 / 2.55 1.86 1.39 |
| Warm semantic edit | 11.38 | 2 | Enabled, warm | 2.13 1.63 1.29 / 2.58 1.74 1.33 |
| Full package and subpackages | 3462.80 | Not split | Bypassed | 0.09 0.67 1.31 / 3.17 1.65 1.27 |

Exact instrument for the first three rows:
`GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=0 ADAMIC_NATIVE_SPLIT=0 ADAMIC_LINT_BENCH=0
XDG_CACHE_HOME=<round-cache> go test ./stage1/cohere/lint -run '^TestRule$'
-count=1 -v -timeout=30m -args -rule eqeqeq`, adding respectively no flag,
`-rule-byte-change`, or `-rule-semantic-change`. Each round has its own empty
cache, reused for its two edits. The author command runs this same test.
The semantic run intentionally exits 1: Node, emitted JavaScript and sanitized
native have byte-identical changed findings, all disagreeing with Go. Both native
ports compile; rejection precedes the owned-mutant execution. All whitespace and
cold checks pass. Object counts include the concurrent correct and mutant builds.

Full instrument: `GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 ADAMIC_NATIVE_SPLIT=0
ADAMIC_LINT_BENCH=0 go test -count=1 -timeout 90m -json ./stage1/cohere/lint/...`.
All five packages pass, including 40 owned mutants. The longer timeout lets the
uncached whole-file builds finish without removing checks. Seven cache-key
mutants drop the adapter, capture filter, manifest/source, port modules, Node
modules, emitted modules, and JavaScript-build modules respectively; each causes
its invalidation test to fail with a stale answer. Registry-selection,
fresh-timing-observation and recovery-marking mutants were also caught.
Selected/full parity covers passing and failing verdicts; a separate split-enabled
run proves byte-identical split/whole-file output for no-var, no-empty and eqeqeq.
Timed throughput probes always execute afresh. Opt-in external-compiler and
throughput benchmarks were not enabled; existing parser recovery limits remain.
Raw logs, traces and benchmark dumps are kept outside the tree.
