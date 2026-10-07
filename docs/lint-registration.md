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

## Checking one rule

From the repository root, run:

```sh
go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -timeout 30m -v -args -rule no-var
```

Replace `no-var` with your directory slug. The convenience wrapper
`go run ./cmd/adamic-lint-check <slug>` invokes the same test command.
Keep the anchors when invoking the test directly: `-run TestRule` also matches
`TestRulesAgree`, which intentionally runs the full corpus.

The check validates the entire discovered registry, snapshots the selected rule
and shared scanner, and retains the five migrated baseline registrations. Other
registrations do not enter the native or emitted JavaScript module graph. Their
module files remain available when the selected rule imports a helper.
It checks every owned witness for a Go finding, the selected rule's captured
upstream tests, and every original inherited corner-case manifest row and option, including `all`
selection, plus selected-rule variants.
A private oracle uses the same restricted registration plus the unchanged Go
baseline for `all` rows. Go cohere, source Node, emitted JavaScript and ASan/UBSan native must agree byte
for byte. The owned mutant runs on witnesses and inherited cases, and must
disagree with Go on each port runtime. Unsupported recovery cases keep the
ordinary harness's explicit refusal check. A phase timing table prints at exit,
including on failure. Existing all-rule tests remain available and unchanged in
scope. `TestSelectedRuleParity` compares restricted and full registration builds
on the same selected-rule corpus for three directories.

Artifacts and observations live below `os.UserCacheDir()/adamic/lint/`:

| Cache | Inputs that invalidate it |
|---|---|
| Go oracle binary | Cohere HEAD, binary working-tree diff, nested submodule identities/diffs, untracked bytes, oracle overlay, generated Go registry, every adapter, Go version, GOOS/GOARCH |
| Upstream capture | Same cohere identity, capture overlay and harness, upstream package, test filter and exact public rule name |
| Go observations | Oracle key, complete manifest fields and every referenced source byte |
| Sanitized native port | Actual imported module names/bytes, generated registry, all `internal/**` bytes, `go.mod`, `go.sum` or its absence, executing compiler identity, native flags, whole/split mode, split job count and runtime library key |
| Emitted JavaScript build | Module graph, compiler source and executable identity |
| Source Node / emitted JavaScript observations | Module bytes, Node version, manifest fields and referenced source bytes |

Harness and Node runtime bytes, execution environment, uid and stack limit also
participate where relevant. Temporary capture destinations and cache location
are transport, and are excluded from the execution identity. Imports come from
the checker's parsed declarations; comments and strings cannot hide dependencies.
Files are published by temporary file and rename. Artifacts publish an entire
completed directory with an integrity digest. Corrupt or partial entries are
misses. Successful observations are compared again on hits, and native still
executes under sanitizers on every check. Builds reject inputs that change while
they are compiling. Content-addressed corpus filenames retain the exact output
protocol across invocations; source snapshots are always written afresh.

`ADAMIC_GATE_UNCACHED=1` bypasses all lint artifact and observation caches, for
both selected and full tests, and calls the original implementations directly.
The existing native runtime library cache retains its own flag-set behavior.
Integration still runs the ordinary uncached all-rule gate before main moves.

`TestLintCacheInvalidation` changes real inputs behind populated caches and
compares their answers to uncached implementations. `TestLintCacheMutants`
removes one key component per cache in an isolated subprocess and requires the
invalidation test to fail with the old answer, including the emitted-build cache.
`TestLintCacheBypassAndIntegrity` checks bypass, byte preservation and corruption.
The optional test flag `-rule-byte-change` appends exactly one space to `rule.a`
or `rule.ts` in the private snapshot for measurement, preserving owned files.
See [the check evidence](../stage1/cohere/lint/check-evidence/REPORT.md) for commands,
measurements, build flags, mutant failures and coverage limits.

The selected-rule fast path uses the opt-in native split compiler with
`Options{Split: true, Jobs: runtime.GOMAXPROCS(0)}`. Correct and owned-mutant
native builds run concurrently. Both retain ASan/UBSan and the ordinary flags.
The compiler caches preprocessed function-group objects under `adamic/units/`;
see [CLANG_UNITS.md](../internal/native/CLANG_UNITS.md) for dependencies and
limitations. A shared declaration or numbering change can invalidate every unit.

The lint harness clears the inherited `ADAMIC_NATIVE_SPLIT` switch and selects
split mode through the API. Full-registration builds and
`ADAMIC_GATE_UNCACHED=1` always use ordinary whole-file compilation.
`TestSplitWholeParity` compares the complete three-rule corpora byte for byte
against a freshly built whole-file native witness, also checking Go, Node and
emitted JavaScript. `TestSelectedRuleParity` separately checks full registration.

For paired measurement, `-rule-whole` uses the ordinary compiler and sequential
correct/mutant builds. `-rule-helper-change` adds a used identity helper to the
private rule snapshot, exercising declaration and numbering changes without
editing the checkout. It can also be passed to `TestSplitWholeParity`.
See [split evidence](../stage1/cohere/lint/split-evidence/REPORT.md) for paired
cold, byte-edit, unchanged and helper-edit timings and actual object counts.

## For rule authors

From the repository root, check your directory slug with one command:

```sh
go test ./stage1/cohere/lint -run '^TestRule$' -count=1 -args -rule <slug>
```

The check validates every descriptor, then checks your owned witnesses, captured
upstream cases, inherited corner cases and options against unchanged Go cohere,
source Node, emitted JavaScript and ASan/UBSan native. It also checks your owned
mutant on all three port runtimes and reports phase timings. The selected native
build uses split compilation; correct and mutant builds run concurrently.

The uncached all-rule run remains the integration witness: every registration,
every upstream package, inherited corpus, owned witnesses and mutants, volume
and count guards. `ADAMIC_GATE_UNCACHED=1` bypasses the lint caches, and those
full-registration native builds use ordinary whole-file compilation. A fast
selected-rule pass does not replace this run. Diagnostic IDs, descriptions,
ranges, fixes and suggestions are part of the byte comparison; changing an ID
must fail even when Node and native agree on the changed result.
