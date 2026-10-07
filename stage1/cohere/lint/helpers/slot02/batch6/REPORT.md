Built three retained .a helpers: stylesheet loadFile, resolveImport and ingestCustomVariant; six consumers each, eighteen dependency entries, zero final blockers.
Commits: claims e283cc2, corrected replacement claims 7767668, implementation 5f03de1; branch codex/lint-helpers-02 on current main e8ba3d5.
Commands: setup 39s/nproc 5; retained isolated helper gate PASS 142.399s; complete helper gate PASS 395.124s; inventory PASS 6.203s; fresh input oracle PASS 1.600s; vet exit 0.
Mutants: eighteen new compiling semantic mutants caught against actual Go on source Node, emitted JavaScript and sanitized native; all thirty-five prior mutants caught again (fifty-three total).
Not covered: whole rules/full repository gate, unavailable external live/corpus gates, invalid string/adapter inputs and panic unwinding; dependency adapters remain explicit.

# Ownership and landing

The published branch was landing-ready at f39cd66 on main e8ba3d5 before any new claim. An explicit wildcard fetch inspected every claims file across all eighteen origin codex/lint-helpers* branches. Shared HELPERS.md reserves the original comments bundle; generic rule-local strict options is not an unclaimed concrete helper. The retained symbols each tie the highest available concrete consumer count, six.

The initial e283cc2 claim overlapped slot 05 for NewUtilityEvaluator and Table.addRepositoryFunctionalRoots. Slot 05 e0b9515 at 03:12:38 UTC precedes e283cc2 at 03:13:26 UTC. Both are withdrawn, with no duplicate implementation delivered or counted. Retain ingestCustomVariant; corrected claims 7767668 were pushed before writing either replacement. evidence/claims.log retains the failed overlap check; evidence/retained-claims.log checks the three actual retained symbols and current main. Only this slot's claim, files and standalone test changed. No shared harness, registration generator, compiler source, main or area branch was edited or pushed.

# Behavior and coverage

loadFile preserves Go's absolute-path resolution, active visiting-stack guard, ordered read/parse/stylesheet append/ingest and visiting cleanup after returned errors. Failed parsing does not append; failed ingestion leaves the already appended stylesheet visible. Existing false visiting entries permit loading and are deleted afterward. Active cycles do not clear the caller's existing true mark. Error formatting retains the exact original cause handle, matching %w wrapping.

resolveImport trims and segments through explicit Go-compatible dependencies, strips all edge quote characters, skips empty modifiers, accepts the source( prefix and refuses every other modifier before resolution. It passes specifier then parent directory to the resolver and preserves wrapped resolver errors. ingestCustomVariant preserves Go Unicode whitespace, the bracket-aware segment call, one trailing -* removal, empty-name skip and map value replacement.

The overlay invokes the unchanged actual private Go helpers; only their dependency call sites are instrumented. Real temporary files and a deleted owned current directory exercise successful loads, missing files, parse/ingest failures, cycles, false marks and an actual filepath.Abs/getcwd failure. Visiting values and key presence, stylesheet append before ingest, parsed-list identity, requested paths, ordered calls, exact messages and nested causes are compared. The resolver is a supplied dependency with success/refusal fixtures. Go stdlib trim/quote/directory results and parser/ingest outcomes are explicit adapter inputs, not implementations claimed by this unit. Want is removed before Adamic reads the corpus.

Observed: 157 actual runtime rule/file/source captures cover every consuming rule. Decoded TypeScript literal/template values and supplementary Go theme/variant/utility test literals produce 1,625 custom-variant params, 1,633 import params and 3,457 real file loads. Real Go, source Node, emitted JavaScript and ASan/UBSan native agree in the retained baseline. Every successful backend exits zero without stderr or sanitizer findings. Consumer sources and coverage regenerate byte for byte; retained-reproducibility.log records both SHA-256 values.

Inferred: readiness.json removes these three prerequisite symbols, eighteen entries across the same six rules. No rule loses its last listed helper blocker. The residual ledger subtracts this slot's eighteen delivered helpers only, assuming the inventory's common AST adapter; it does not assume other workers have landed. RULES.md lists every consumer. These are dependency handoffs, not observed integrated rule findings/fixes.

# Commands and evidence

All test output was written directly to logs, never piped. Source /workspace/adamic-tools/env.sh first.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s; build cache warm 39s; done 39s on 5 processors. nproc: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0.
- python3 .../batch6/testdata/regenerate.py > evidence/retained-regeneration.log 2>&1: all six consumers captured, 157 rows. The capture protocol accepts only the eight known upstream external failures listed below. This is not a passing integrated Go rule-package gate.
- Retained capture rerun: evidence/retained-repro-capture.log; source/coverage byte equality: evidence/retained-reproducibility.log, PASS.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch6$' -count=1 -v -timeout=15m > evidence/retained-isolated.log 2>&1: PASS 142.399s, all eighteen new mutants caught.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 395.124s; all fifty-three compiling semantic mutants caught.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 6.203s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle-final.log 2>&1: PASS 1.600s, all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet-final.log 2>&1: exit 0, empty log.

# New semantic mutants

Each must compile, run all three Adamic backends successfully, agree across them, and only then disagree with actual Go. A compile refusal, crash or sanitizer failure is not credited. Baseline files remain unchanged; mutants use temporary copies.

| Mutation | Actual Go witness |
|---|---|
| Omit absolute-path error cause | Deleted cwd preserves getwd and syscall causes. |
| Omit parse-error cleanup | Parse failure deletes both the value and key of the active visiting mark. |
| Accept empty import specifier | Go refuses before invoking the resolver. |
| Retain visiting mark after ingestion | Successful load clears its active mark. |
| Set visiting false during ingestion | Go ingestion observes a true active mark. |
| Append stylesheet after ingestion | Go ingestion already sees the appended path. |
| Omit read-error cause | Missing file preserves the PathError and syscall cause chain. |
| Omit read-error cleanup | Missing file clears the mark before returning. |
| Suppress cycle guard | Go refuses an active cycle without reading the file. |
| Keep edge quotes in specifier | Go resolver receives foo.css without its quote characters. |
| Refuse source(...) modifier | Go accepts and invokes the resolver. |
| Accept unsupported modifier | Go refuses theme(reference) without invoking the resolver. |
| Omit resolver-error cause | Go preserves fixture resolver unavailable as the original cause. |
| Swap resolver arguments | Go receives specifier first and parent directory second. |
| Exclude NEL from whitespace | Go strips U+0085 around dark-*. |
| Include BOM in whitespace | Go retains U+FEFF in the name and segment input. |
| Keep functional suffix | Go registers dark, not dark-*. |
| Register empty name | Go leaves the map unchanged for an empty name. |

The thirty-five prior semantic mutants are also rerun in the complete package gate. Their witnesses remain documented in ../landing/REPORT.md and the individual batch reports; the final log records each mismatch. They include all four inherited option/policy mutants, three first-batch name/cache mutants, four second-batch namespace/fields/whitespace/alias mutants, five factory/math mutants, five prefix/underscore mutants, four utility lookup mutants, four root-scan mutants and six public-sort value/header/alias mutants. Earlier first-batch tests compare Go/source Node/native; later batches additionally compare emitted JavaScript. No extra emitted-JavaScript coverage is claimed for the original tests.

# Superseded attempts and limits

Initial overlap work had an array-index type refusal and unsupported nullable-object representation, then a duplicate-registration mutant survived because the fixture's first and last registrations pointed at the same definition. The attempted fixture correction still did not distinguish those pointers. superseded-overlap-gate.log is therefore a failed package run, not final evidence. Those two ports were withdrawn and deleted for ownership; no validation claim or implementation for either is delivered here.

The first retained overlay build had a leftover withdrawn exporter and an unused os import after wrapping the filesystem call. Owned scaffolding was corrected. Stage 0 refused shorthand function values and a captured nested function in the driver; explicit callback wrappers and a top-level function with a state object compile and pass. The environment then restarted during the retained mutant run; interrupted-environment-run.log has no completed result and is not credited. Files and published claims persisted. The retained isolated run and fresh filtered oracle were rerun after the restart. Every superseded log is retained separately.

Eight unavailable Tailwind live/corpus gates remain outside this unit: TestUnknownClassFixturesActuallyRan, TestConflictFixturesActuallyRan, TestCanonicalClassesPlacementIsAccountedFor, TestConflictingClassesPlacementIsAccountedFor, TestClassOrderFixturesActuallyRan, TestCanonicalFixturesActuallyRan, TestUnknownClassesPlacesEveryCorpusClass and TestClassOrderLiveMatchesTheEngineOverTheCorpus. Capture observes inputs before those external gates and rejects any unexpected failure or missing consumer.

No full repository gate or integrated rule findings/fixes/suggestions parity is claimed. Filesystem/path, segment, CSS parser/ingester, Go quote/trim and error/node arena adapters remain externally owned prerequisites. Bounded fixtures do not prove arbitrary repository graphs. The native loader does not itself traverse imports; it invokes its explicit ingestion dependency. Dependencies must return normally using error handles; panic unwinding, nil receivers/maps, malformed adapters, invalid UTF-8 bytes and unpaired UTF-16 surrogates are outside the tested projection. The batch leaves all six rules with residual blockers.
