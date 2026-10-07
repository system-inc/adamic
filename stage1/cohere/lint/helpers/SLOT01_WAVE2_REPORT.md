Built: NewClassLiteralReader, classValuesIn and classLiteralFrom, one helper per .a file.  
Commits: implementation c986bcc; claims c250b83, 3cc9df4/58dc202 and e7eff58 pushed before code.  
Commands/results: complete helper package PASS, 115.500s; go vet ./... clean; filtered uncached input oracle PASS, 1.036s.  
Mutants: eight new compiling semantic mutants caught by actual Go output comparisons; six previous and four inherited mutants rerun.  
Not covered: native regex provider, whole-rule findings/fixes, dynamic fixtures, cross-slot integration and full repository test gate.  

# Slot 01 second batch report

Branch codex/lint-helpers-01; base and pinned Go cohere match the first report. Everything previously committed was pushed at the start (already up to date). Every origin codex/lint-helpers* branch was fetched and every claims file inspected before each new claim. No compiler ownership file, rule entry point, frozen readiness ledger or cohere source was edited. No PR was opened.

## Delivered behavior and consumers

| Helper | Remaining consumers | Final blockers removed alone |
|---|---:|---:|
| NewClassLiteralReader | 12 | 0 |
| *ClassLiteralReader.classValuesIn | 11 | 0 |
| classLiteralFrom | 11 | 0 |

This batch removes 34 dependency entries for 12 distinct Tailwind rules; it adds no fully helper-ready rule by itself. The factory serves every rule below. The memoizer and literal constructor serve the same list except no-concatenated-classes. Exact symbols, per-helper consumers, residual dependencies after this batch and after all six slot 01 helpers are in slot01_wave2_readiness.json. Readiness is conditional on the inventory's common AST adapter and supplied Go-equivalent dependencies, and does not assert rule implementation.

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

The factory preserves deduplicated name maps, ordered valid compiled patterns (including duplicates), skipped invalid patterns and unbound state. Its regex compiler is an explicit callback; oracle fixture pattern objects are test data, not a delivered native regex engine. Stage 0 refused undefined/null-valued APIs, so the delivered representation uses valuesBound=false and a tagged {valid, pattern} compiler result. Unsupported provider features must be explicitly refused by the provider.

The memoizer preserves unbound fresh reads, bound sharing by exact node identity, cached empty/nil-node results, and distinct-node isolation. Its external readClassValues dispatcher remains another worker's helper. The literal constructor preserves node/text/origin, zero edges and the real token range excluding quotes when its length is at least two. Parser token ranges and decoded text are explicit adapter inputs; no offset is inferred from text or trivia-inclusive Loc. API contracts are in SLOT01_WAVE2_README.md.

## Commands and observations

All test stdout/stderr went directly to files. All baseline and semantic mutant processes must exit successfully with empty stderr. Compiler refusal, panic and sanitizer failure do not count as mutant catches.

- source /workspace/adamic-tools/env.sh in build shells; nproc printed 5.
- go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/slot01-wave2/final.log 2>&1: PASS, 115.500s. Entire touched package, 15 top-level tests, including previous and inherited helpers/refusals.
- go vet ./... > /tmp/lint-helpers-01-wave2-vet.log 2>&1: clean, empty log; retained as evidence/slot01-wave2/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave2-oracle.log 2>&1: PASS, 1.036s; six fixtures, zero probe hits and six misses; retained as oracle.log.
- gofmt -l stage1/cohere/lint/helpers and git diff --check: empty output.

No setup rerun was necessary in this continuation. The settled-tree setup from this same unit completed in 75s on five processors; its full timing lines and first failed/retried attempt are recorded in SLOT01_REPORT.md and evidence/slot01/setup.log. Timing lines: go ready (0s); clang ready (1s); node ready (1s); submodules ready (1s); build cache warm (75s); done in 75s, cgroup cpu.max 400000 100000, 17.6 GB.

Factory: all 12 consumer fixture files contribute 1,016 extracted strings, plus controls = 1,936 batches. Actual Go private reader state and pattern behavior match Node and ASan/UBSan native for 148,474 output lines. The provider table is independently generated with Go regexp.Compile, then the expected output comes from the actual NewClassLiteralReader result.

Memoizer: all 11 consumer fixture files contribute 956 extracted strings, plus controls = 1,876 batches. Each string is embedded as the value in two separate parsed JSX attribute nodes. Repeated and nil nodes are queried with bound and unbound readers, and one result slice is deliberately mutated after observation to expose sharing. The actual Go classValuesIn result and cache size match Node/native for 37,832 output lines.

Literal constructor: the same 11 consumer files contribute 956 extracted strings, plus controls = 1,888 parser batches. Every parsed string/no-substitution-template node is queried with five origins. Trivia, escape, Unicode and unterminated controls exercise real parser TokenRange and short ranges. Go matches Node/native for 5,940 output lines. These are helper corpora from static string expressions, including fixture labels/messages, not execution of whole consuming-rule test harnesses. Dynamic fixture construction is not evaluated.

## Every new mutant

| File | Compiling mutation | Go observation that caught it |
|---|---|---|
| tailwind_new_class_literal_reader.a | reverse valid compiled patterns | line 14: mutant ^custom$, Go .*[Cc]lassName$ |
| tailwind_new_class_literal_reader.a | accept invalid patterns | line 1: mutant 6 patterns, Go 5 |
| tailwind_new_class_literal_reader.a | store false for attribute names | line 2: mutant false true, Go true true |
| tailwind_class_values_in.a | bypass cache lookup | line 14: mutant empty string, Go ! after shared result mutation |
| tailwind_class_values_in.a | cache unbound reads | line 1: mutant cache size 1, Go 0 |
| tailwind_class_values_in.a | omit empty cache entries | line 19: mutant cache size 2, Go 3 |
| tailwind_class_literal_from.a | leave opening quote in range | line 2: mutant start 225, Go 226 |
| tailwind_class_literal_from.a | strip ranges shorter than two | line 522: mutant 1..0, Go 0..1 |

The six earlier slot 01 mutants were rerun: remove path normalization (line 7), omit class expressions (13400), omit class default (1), share each of the three default arrays (3). Four inherited mutants were also rerun: raw-control acceptance (15157), overlapping oneOf acceptance (15166), missing message interpolation (22166), unknown-field acceptance (15178). Their Go comparisons caught each successful compiled mutant; details are in the full final.log and first report.

## Claims and limits

Factory claim c250b83 at 00:43:26 UTC preceded slot 05 cba50e2 at 00:43:32. Slot 05 subsequently withdrew the duplicate. Memoization claim 3cc9df4 was corrected to readiness's pointer-receiver symbol in 58dc202 before its code was written. Literal construction claim e7eff58 was pushed before code. Every helper was tied for or held the highest unclaimed consumer count when selected; the comment bundle remained excluded under the shared branch's earlier HELPERS.md claim.

The first factory API compile attempts exposed stage 0's unsupported undefined/null lowering; no compiler file was edited or rejection credited as a mutant. Early oracle harness issues (relative Go parser path, Go-vs-JavaScript JSON HTML escaping, and numeric console typing) were corrected before successful baselines. Delivered comparison outputs use raw text and numeric summaries, preserving semantics across hosts.

Not covered: production Go-equivalent regex compilation/execution, cross-slot callback/arena integration, template payload content mutation (the generic cache preserves both lists but the focused memo corpus mutates literal payloads), whole-rule findings/spans/fixes, dynamic fixture construction, arbitrary settings or ASTs beyond these corpora, and the full repository go test ./... gate. Required bounded validation ran the entire touched helper package, repository-wide vet and filtered uncached external oracle. This batch adds helper prerequisites, not executing native rule listeners.
