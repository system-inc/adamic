Rebuilt the step 16 scout as documentation, measurement tools and 15 fixture contracts on main; no new compiler implementation.
Base 7a10c877667582a15326acb10fa22fa7a0c45fb8; source f3a560a7b74676712c2bfd750473dd2013ada64a; delivery compiler/generics-scout-main.
Build, internal vet, changed-file a-check, stage1 Gap/Gaps/Probes, stage3 fixtures, focused uncached oracle and regenerated counts pass.
Nine retained mutants are caught; the five historical mapper mutants were rerun, with three surviving and retired with the redundant mapper addition.
Full tsc compilation remains checker-blocked; the next generic source replay remains unverified because the census snapshotter panics on argumentFacts.

## Change and conflict resolution

Applied the source branch's binary net diff from merge base 8cb5e7c189fc431ab3cab7e0be2c0530e722a598, without merging its old ancestry.
The only three-way conflict was internal/oracle/counts.md. Kept every main row, added only the scout rows, then regenerated the table.
The generic.go patch merged without conflicts and preserved main's newer optional and union inference. All 15 updated outcomes also pass with that addition removed, using an overlay of the unmodified main file. Omitted the redundant mapper path rather than carrying additional reflection and guards whose original positive catchers no longer fail. Main already supplies the scout's admitted shapes.

Fixture 13 now stops on a local value of type T["value"], rather than its optional return. Fixture 14 now lowers and matches Node. Added refusal headers to fixtures 06, 07 and 08 for the changed-file a-check; updated fixture 14's historical capability comment. landing-records.json preserves original and current hashes and current source Node observations. The historical fixture and census ledgers are unchanged.

## Wall contracts

| Fixture | Current outcome |
| --- | --- |
| 01_identity | Lowered |
| 02_optional_return | Lowered |
| 03_callback_return | Lowered |
| 04_function_value | NotYet: a generic function as a value |
| 05_nested_class | Lowered |
| 06_polymorphic_recursion | Refused: polymorphic recursion |
| 07_generic_cast | Refused: a cast the runtime can't check |
| 08_readonly_view | Refused: readonly field pos becomes writable |
| 09_recursive_optional | Lowered |
| 10_identifier_multimap | Lowered |
| 11_indexed_result | NotYet: a function returning T["value"] |
| 12_mixed_indexed_result | NotYet: a function returning U or T["value"] or undefined |
| 13_constrained_local | NotYet: a value of type T["value"] |
| 14_optional_literal_union | Lowered |
| 15_array_callback | Lowered |

The admitted shapes cover enclosing binders, optional number/boolean/string results including falsy values, callback results, concrete class argument layouts, recursive optional calls, and the checked array/callback reduction of core.ts:67. These are eight executable reductions, not acceptance of the full compiler corpus. Generic body relations belong to compiler/generic-body-relations (#js89dcw); they are neither duplicated nor marked complete here.

## Commands and evidence

Source /workspace/adamic-tools/env.sh; GOPROXY=https://proxy.golang.org|direct. Each command's complete output is a log, and exit statuses are recorded in landing-records.json.

| Command | Result / log |
| --- | --- |
| bash cloud/setup.sh | Initial failure: no space left on device in Go build cache. go clean -cache freed 29 GB; retry passed, /tmp/generics-scout-setup-retry.log |
| go build ./... | Exit 0, /tmp/generics-scout-final-build.log |
| go vet ./internal/... | Exit 0, /tmp/generics-scout-final-vet.log |
| changed .a against origin/main, adamic c with gate refusal headers | 15 pass: 12 checked, 3 pinned refusals, /tmp/generics-scout-a-check/results.json |
| stage1 Gap/Gaps/Probes tests | Exit 0, /tmp/generics-scout-stage1.log; actual regexp is Gap or Gaps or Probes |
| go test ./stage3/fixtures -count=1 -v | Exit 0, 33.758s, /tmp/generics-scout-stage3.log |
| Focused uncached step16 and generic fixture oracle | Exit 0, 2.199s, /tmp/generics-scout-final-own.log; actual regexp uses an unescaped alternation |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts | Exit 0, 47.222s, /tmp/generics-scout-final-counts.log |

The exact selectors are:

```sh
go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestStep16|^TestNativeAgreesWithNode$/stage3/fixtures/generics/' -count=1 -v
```

The oracle checks source Node, JavaScript backend, release native, ASan/UBSan and leak checks for the eight supported fixtures. All fifteen sources were also rerun on Node with node:module.stripTypeScriptTypes; their stdout, stderr and exit match the historical observations. NotYet and Refused cases have no claimed native execution.

Setup timing lines: Go ready 0.023s; Node ready 0.025s; markdown skip 0.009s; submodules ready 0.084s; markdown ready 0.085s; clang ready 0.201s; go build ready 375.906s; test binaries deferred 376.203s; cache warm 376.207s; done 376.332s. nproc=5, cgroup quota=4 CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8.

## Mutants

| Retained check / mutant | Observed catcher |
| --- | --- |
| TestStep16IdentityMutant: numeric identity result becomes 17 | Both backend stdout differ from source Node; valid C and clean sanitizers/leaks |
| TestStep16OptionalResultMutant: present number becomes undefined | Both backend stdout differ from Node; valid optional ABI and clean sanitizers/leaks |
| Same optional result mutant on fixture 15, test-only Go overlay | Both backend stdout differ from Node; /tmp/generics-scout-array-mutant/test.log |
| audit.py --mutant | Unique-root assertion, exit 1 |
| audit.py --mutant-selection | Scoped-binder selection assertion, exit 1 |
| audit.py --mutant-markdown | Refusal-table parser assertion, exit 1 |
| parameters_test.py --mutant | Enclosing-binder assertion, exit 1 |
| resume.py --mutant | Changed-overlap assertion, exit 1 |
| retirement.py --mutant | Moved-site retirement assertion, exit 1 |

The corresponding normal audit, parameters, resume and retirement tests exit 0. No build or compiler-warning failure is credited as a mutant catcher.

All five original compiler mutants were independently rerun with Go overlays on the candidate mapper patch. mapper-disabled, identity-skipped and union-size-removed exit 0: their old catchers no longer fail on main. optional-gating-removed and constraint-guard-removed exit 1 after compilation: respectively fixtures 11 and 13 incorrectly become accepted. These results are preserved in landing-records.json and /tmp/generics-scout-mutants/. The mapper patch is absent from the final rebuild, so all five mapper-specific checks are retired, not claimed as newly certified implementation checks. The final rebuild's nine retained mutants all fail their comparisons or assertions as intended.

## Counts rows

A/F/R/L/P/G are allocations, frees, retains, releases, peak live and region values. These eight rows are additions for executable fixtures. No pre-existing row changes; no performance or retirement credit is claimed.

| Added fixture | A/F/R/L/P/G | What the new row measures |
| --- | --- | --- |
| 01_identity | 3/3/3/7/3/0 | Runtime-built string through identity and enclosing relay |
| 02_optional_return | 5/5/3/9/3/0 | Present falsy optional results, owned string and absent fallback |
| 03_callback_return | 8/8/3/11/3/0 | Callback result, enclosing relay and owned string |
| 05_nested_class | 8/8/5/16/7/0 | Nested generic class instances with distinct concrete layouts |
| 09_recursive_optional | 3/3/1/4/3/0 | Recursive optional number/string results |
| 10_identifier_multimap | 11/11/27/36/9/0 | Generic map subclass, array values and factory relay |
| 14_optional_literal_union | 3/3/0/3/3/0 | Number/zero/undefined result and printed driver |
| 15_array_callback | 18/18/7/27/4/0 | Optional readonly arrays, checked element access and callback results |

## Source stop and ruling questions

Prepared pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8 with bash stage3/apply.sh /tmp/generics-scout-adapted (exit 0). This is the current adaptation, not the historical scout's exact corpus. The ordinary command go run ./cmd/adamic c /tmp/generics-scout-adapted/src/tsc/tsc.ts exits 1. Its first diagnostic is src/compiler/builder.ts:1246:69, TS2345: Path | undefined is not assignable to string. This is a checker stop before lowering.

The generic source selector core.ts:220:1 (contains, the generic comparator default shape) was attempted with go run ./stage3/census/latent/replay -project /tmp/generics-scout-adapted/src/compiler -where /tmp/generics-scout-adapted/src/compiler/core.ts:220:1 -kind NotYet -reason 'a generic function as a value'. It exits 1 because the census snapshotter reports `latent state copy: unexported IR field argumentFacts`. No generic diagnostic was reproduced. This is a tool blocker, not evidence that the generic wall disappeared. The next generic source stop remains unverified; the reduced function-value fixture still pins the reason. The snapshotter needs to handle the current IR's argumentFacts before that measurement is possible. No unrelated census fix is included.

The scout's unresolved proposals remain proposals. For the specialization depth ceiling, propose a configurable resource diagnostic rather than a language refusal. For generic function values, propose preserving JavaScript function identity while specializing only proven concrete call signatures; retain NotYet until the identity/dispatch representation is ruled and implemented. No code is built past either choice. The dependent-body questions have already been ruled and are assigned to compiler/generic-body-relations.
