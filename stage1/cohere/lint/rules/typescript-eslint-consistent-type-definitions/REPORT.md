Built: moved the batch rule into an owned .a descriptor directory; certification is blocked by the shared multiple-fix guard.
Commit: the branch commit carries the rule and evidence; its SHA is reported with the push.
Checks: lint-registry, gofmt and vet pass; TestOwnedWitnesses, TestRulesAgree and the assigned TestMutants subtest fail at the unchanged Go oracle's fix-shape guard.
Mutant: the message mutation is defined and valid source; it is not certified caught because the oracle fails before comparisons.
Not covered: complete findings/fixes/suggestions parity, mutant execution and sanitized runtime agreement cannot be certified on this harness.

The multi-edit guard was removed upstream; the current rerun and remaining fixer mismatch are documented in MULTIEDIT_REPORT.md.

## Source and registration

Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898.
Batch source: origin/codex/stage1-lint-batch4-typescript 63782c53678711717f9fd4f12762ce3d103b9a3d,
stage1/cohere/lint/consistent_type_definitions.ts, named by DEDUP_LEDGER.md.
Messages moved verbatim into messages.a. The descriptor listens to TypeAliasDeclaration
and InterfaceDeclaration, uses node:true, and its visit receives the node then its index.
No shared registration, oracle, context, compiler or corpus file was changed.

The Go adapter returns the unchanged upstream rule and decodes captured field 5
into ConsistentTypeDefinitionsOptions with upstream defaults. It accepts the
upstream bare string schema through DecodeConsistentTypeDefinitionsOptions and
the decoded struct emitted by capture. It never drops supplied options as nil.

upstreamTest is TestConsistentTypeDefinitions, covering all six real functions:
TestConsistentTypeDefinitionsMatchesUpstream,
TestConsistentTypeDefinitionsRewritesWhatUpstreamRewrites,
TestConsistentTypeDefinitionsDeclinesInsideDeclareGlobal,
TestConsistentTypeDefinitionsDecoderDefaultsToInterface,
TestConsistentTypeDefinitionsKeepsHeritageOrder,
TestConsistentTypeDefinitionsKeepsTheDeclarationModifiers.
The corpus includes both styles, parenthesized object types, heritage order,
exports/default exports and deliberate no-fix global augmentations.

The local repair helper composes the batch rule's edit spans into the current
single-edit Finding. It preserves replacement-source behavior but cannot promise
byte-identical multi-edit findings. A shared multiple-edit Finding and wire model
is needed for that certification; this unit does not redefine or bypass the oracle.
The default witness needs no option sidecar and reports an upstream finding.

## Exact blocker and reproducer

Witness: type Shape = { value: string };
Default style: interface. Go's consistentTypeDefinitionsAliasFixes returns three
automatic edits: replace the keyword, replace the equals/prefix span with one
space, and remove the trailing semicolon. stage1/cohere/lint/testdata/oracle.go:74
requires len(d.Fixes)==1 and panics with unexpected fix shape. Interfaces under
the type option also have multiple automatic edits. The shared RuleContext and
Finding likewise carry one automatic edit. This is not the new options guard:
option decoding succeeds, and the failure is the automatic-fix shape.

Reproduce from the repository root:

go run ./cmd/lint-registry
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$|^TestMutants$/object-alias-message-changed$' -count=1 -v -timeout 30m > /tmp/consistent-type-definitions.log 2>&1

Final gate exits 1. All three selected tests fail with the oracle's unexpected
fix shape; source compilation succeeds before that blocker. No skip, weakened
expectation, recovery flag or adapter transformation was introduced. Shared
source remains untouched. The declared mutant changes only the interface message
and would require full Go/Node/emitted-JavaScript/native comparisons to certify it.
Its current failing test is not counted as a killed mutant.

## Commands and logs

bash cloud/setup.sh: pass, total 48s; Go 0s, clang 0s, Node 0s,
submodules 1s, cache warm 48s. nproc: 5; CPU quota: four cores; memory: 17.6 GB.
Environment: source /workspace/adamic-tools/env.sh.
go run ./cmd/lint-registry: pass; registry.log lists the rule.
gofmt -w then gofmt -l on this directory: pass, empty formatting log.
go vet ./...: pass, empty vet log.
Initial separate required tests and the final scoped rerun are preserved under
evidence. Initial implementation failures are historical: the visit argument
order and selected-rule option isolation were corrected, and bounded slices
replaced a startsWith position overload the compiler does not implement.
gate-final.log records the remaining genuine blocker on the corrected source.
No full repository gate or native-versus-Go timing was run because the complete
comparison cannot run. Only the assigned branch is pushed; no pull request.
