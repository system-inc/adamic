# Parallel work

A fixture owns its source, its oracle options and its counts. A lint rule should own
its implementation, registration descriptor, options adapter, messages and probes.
Adding either must not require appending a name to another worker's file.

## Oracle fixtures

Add `name.a` (or `.ts`) and `name.a.oracle.json` beside it under
`internal/oracle/testdata`. Discovery recursively globs `*.oracle.json`, sorts by
repository-relative source path, and runs only those entry points. Imported helper
modules and specialized probes such as `weak/`, `fresh_refused/` and
`killed_after_output.a` have no sidecar. This distinction preserves today's set;
a glob of all source files would wrongly run these programs independently.

An ordinary fixture's sidecar is:

```json
{
  "lowers": true,
  "checked": false,
  "input": false
}
```

All three booleans are required. `lowers: false` requires a `lower.NotYet` refusal
with where and what. `checked: true` means an inserted check fires: native and the
JavaScript backend must agree, native must exit 70, and source Node must run on.
Unknown fields, missing booleans, missing sources, duplicate source registrations,
invalid paths and inconsistent input options fail loudly.

An input fixture uses `input: true`, `lowers: true`, `checked: false` and may add
`arguments` (a JSON array of strings), `unreadable: true` (a forbidden file added as
the last argument) or `writes: true` (a private writable directory prepended).
Use `argumentsHex` instead of `arguments` for byte-exact argv, including malformed
UTF-8; each string is the hex encoding of one argument, and `""` is an empty
argument. The migrated `arguments.a` keeps every byte from the original registry.

`uncounted: true` excludes nondeterministic counts, currently only
`stack_overflow.a`; it does not exclude execution. External examples have one
reference sidecar each under `testdata/examples`, with a `path` naming their
existing repository-relative source. Their sources and imports have not moved.

Record counts for your fixture from the repository root:

```sh
go test ./internal/oracle -run 'TestCountsAreRecorded/fixtures/internal/oracle/testdata/name.a$' -count=1 -args -update-counts > /tmp/name-counts.log 2>&1
```

Read the log and commit `name.a.counts.json` with your source and options. A filtered
update writes only fixtures that ran, and an unchanged measurement writes nothing.
A failed sibling does not discard already successful independent updates; review
the resulting files and the log. No worker allocates a registry slot or edits a
shared table. New count files omit `order`; they sort by source path after migrated
rows. Existing `order` values preserve the old table's presentation and stay fixed.

For humans, without running the compiler:

```sh
go run ./cmd/oracle-fixtures -list
go run ./cmd/oracle-fixtures -counts > /tmp/counts.md
```

The second command renders the same header and rows that used to live in
`internal/oracle/counts.md`. Do not check in the rendered table as a new shared
baseline. A missing count file makes rendering fail, rather than silently lose a
fixture. The counts test compares every measured row even when observations were
cached. `ADAMIC_GATE_UNCACHED=1` still bypasses all observation caches.

The generated-C cache key and runtime/toolchain keys are unchanged. The harness
identity now includes `fixturedata/*.go`. An input probe also includes its own
sidecar bytes, so changing its arguments or permission options invalidates its
saved verdict. Sidecars for unrelated fixtures and count files are not added to
the shared cache identity. Ordinary Node/native observations remain evidence;
the fixture options determine comparisons again on every run.

## What registration does on the scanner branch

Inspected `origin/codex/typescript-scanner` at
`0090256e607c3f2de7d5b67cef680ec010f95c1d`. This section describes that commit;
the proposal below has not been implemented on that worker's branch.

There is no directory registry in the Adamic port. Registration is executable
dispatch spread across these files:

- `stage1/cohere/lint/main.ts` reads a manifest row's rule name (field 1), equality
  mode/null policy (fields 2/3), empty-catch option (field 4), JSON options (field 5)
  and recovery marker (field 6). It constructs `Settings` and `Linter`.
- `lint.ts` defaults an empty selection to `all`. `enabled(name)` accepts `all` or
  that exact name. `run()` parses, allocates the parent index table, constructs
  `VolumeRules`, calls its `prepare(root)`, walks and sorts findings.
  `walk()` dispatches the original five rules directly by node kind, calls
  `additional(index, parent)` for the remaining original syntax rules (including
  file-level BOM and warning-comment work), calls `VolumeRules.visit(index)`, then
  recurses into children. Rule methods and message imports live in these shared
  files; adding a rule currently extends these conditionals or methods.
- `volume.ts` supplies the additional ten rules. `visit()` calls `visitSecond()`
  and its own kind/selection guards. It shares parser, scanner, parents, findings
  and settings with `Linter`. `prepare()` fills ancestry before visitor checks.
  `settings.ts` parses the restricted JSON option grammar into a lowercase-keyed
  map; it does not discover rules. `messages.ts` and `volume_messages.ts` are
  shared message sources.
- The independent Go oracle's actual explicit rule list is
  `stage1/cohere/lint/testdata/oracle.go`, `collect()`, line 115: a 30-element
  `[]rule.Rule` from cohere's core, base, nexus, adamic and typescript packages.
  It filters by manifest rule name, constructs `rule.Context`, decodes the
  rule-specific typed options through a switch and individual conditionals, calls
  `subject.Run(ctx, options)`, collects listeners and invokes listeners by AST
  kind during a preorder walk. It stable-sorts diagnostics by source position.
  `lint_test.go` builds this file inside cohere through a Go overlay, so cohere's
  internal packages remain accessible.
- `lint_test.go` also has a shared nine-file `portFiles` list for copying and
  mutating the port. `testdata/volume_rules.json` is a separate 20-row corpus and
  frequency work list, not the dispatch registry; it includes rules not yet in
  the port. These lists would also conflict if every rule worker appended to them.

## Directory registration proposal for lint workers

Do this once on the scanner branch, then give each worker one directory such as
`stage1/cohere/lint/rules/no-debugger/`. For names containing `/` or `@`, use a safe
directory slug and keep the exact public rule name in `rule.json`. Each directory
owns `rule.ts`, `messages.ts`, `rule.json`, `oracle.go`, `testdata/` and its mutant
specification. Shared parser, finding, settings and traversal code are infrastructure
owned by the integration worker.

The descriptor names the public rule, module and named factory export, interested
node kinds, optional prepare/file/finish hooks, and corpus provenance. Its Go
adapter exports the upstream `rule.Rule` value and an options decoder for that
rule, preserving existing typed defaults, equality mode/null and catch options.
Keep corpus frequency/provenance in this descriptor rather than a shared JSON array.

Use a deterministic Go generator that recursively globs `rules/*/rule.json`.
Reject duplicate public names, bad exports, missing adapters and unsupported hook
kinds. Generate named TypeScript imports and static factory/visitor dispatch into
an ignored build directory. Adamic compiles these imports normally; native code
cannot discover TypeScript modules dynamically at runtime. Generate bucketed
node-kind dispatch so fifty rules do not imply fifty unrelated checks per node.
A rule context should carry source, parser, scanner, immutable parent indexes,
settings and the findings destination. Rules must not point back to their linter;
this keeps the reference graph acyclic. Construct listeners in generated code
with explicit types and named imports, without casts or unchecked reflection.

Generate the Go oracle selection and adapter calls from the same descriptors,
but run the unmodified upstream cohere rules. Extend the existing overlay to
include the per-rule Go adapters and generated selection source in its virtual
package, and build all those files. Merely adding adapters to the filesystem
would not work with today's single-file `go build`. Keep option decoding in each
adapter rather than adding cases to `oracle.go`.

The infrastructure tests should derive copied port files from the discovered
module tree, replacing `portFiles`, and derive corpus rules from descriptors,
replacing `volume_rules.json`. A rule worker adds only their directory. Run the
generator before every build/test; verify regeneration is deterministic; keep
its output ignored and never commit a generated registry that fifty branches
will all change. CI must discover and test each descriptor and fail on malformed
ones, even if a particular corpus filter does not select that rule.

Preserve current prepare-before-walk behavior, preorder traversal, exact public
names/options, finding tie order (`no-labels` first at equal start in the port),
fix conflict ordering, repeated fix passes and recovery refusal. Pin legacy tie
order in each migrated descriptor if needed; new descriptors use lexical name
order, without a shared ordinal allocator. Before adoption, compare all current
Node/native/Go outputs and fixes, then mutate one descriptor to omit its listener
and prove its own upstream corpus fails. Also prove a malformed descriptor and
a duplicate name fail generation. Until this infrastructure lands, workers
should keep each rule's implementation and tests isolated and let its owner
integrate dispatch once; this document does not change registration there.

## Migration evidence

`parallel-work-evidence/fixtures-before.json` lists the actual Go registrations
captured on main `5d4c801`, including both init appenders and the input registry.
`fixtures-after.json` comes from the discovery command. Both are sorted by source
path for comparison; the arguments are hex so their bytes survive JSON. These
are historical evidence, not maintained registries. Both contain 263 fixtures:
257 ordinary and 6 input. `counts-before.md` is the old 262-row table, and
`counts-after.md` is rendered from the independent files. Both diffs are empty.
See `REPORT.md` in that directory for gate, cache and mutant commands and results.
