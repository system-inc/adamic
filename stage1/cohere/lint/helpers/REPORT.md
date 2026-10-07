# Lint helper unit report

Four Adamic TypeScript helpers remove the final listed helper blockers for **46 of the frozen 198 rules**: 30 through the option prerequisites, then 16 through policy rendering. This is readiness for rule workers, assuming the inventory's common AST adapter, not 46 completed lint rules. The other 152 keep their remaining dependencies in `readiness.json`.

## Branch and claim

- Branch: `codex/lint-helpers`.
- Fetched main base: `5d4c8012a0877094134e6c6bac367ff68f9313e8`.
- Inventory merged: `73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf`.
- Merge commit: `d3d1b50576ba8f7a8ed1a6069be0ec14539ce171`.
- Claim commit: `acb5e0f405a642d0a2bfb7b780630caf1c74a154`, pushed before implementation. Its projection was 48; the measured handoff is 46 because boundaries and no-extra-boolean-cast still need custom option work.
- Implementation and evidence are committed on this branch; the final response names their commit. No pull request is opened.

## What was built

Each helper owns one file. `OptionsJson` reads strict JSON into a flat numeric-index arena, retaining numeric lexemes and Go's decoded strings without ownership cycles. `OptionSchema` checks the supported upstream schema vocabulary, local references, composite alternatives and deep equality. `StrictOptions` checks Go target type descriptors, exact tagged key spelling, ASCII untagged folding, nested shapes and integer bounds through signed 64-bit. Integer range checks use decimal text, including values above JavaScript's exact numeric range. `PolicyMessage` renders Go's already-resolved catalog, selects object phrases before sorted value substitution, and refuses missing/extra values and invalid selections.

The Go overlay exports the real resolved policy catalog and discovers real registered option target types. Custom JSON/text decoding is explicitly unsupported. Regeneration changes no cohere source. `catalog.json`, `descriptors.json`, deduplicated `cases.json.gz` and `coverage.json` reproduce byte for byte. See `README.md` for the API contract and limits, and `../HELPERS.md` for all 30 and 16 rule names.

| Helper | Consumers covered | Additional complete helper readiness |
|---|---:|---:|
| JSON reader | prerequisite for 103 option consumers | 0 |
| Schema validator | 94 complete upstream schemas | 0 before strict target validation |
| Strict target validator | 97 registered target types | 30 with JSON and schemas |
| Policy renderer | all 54 policy consumers | 16 after options |

These are overlapping consumer counts. A helper-ready rule can still need its own defaults, custom configuration conversion, complete AST adapter, listeners, reports, fixes and suggestions. No inventory stage1 status is rewritten.

## Validation commands and observed output

All test commands wrote logs directly, without pipes. Environment: Go 1.27.1, clang 20.1.8, Node 24.19.0; `source /workspace/adamic-tools/env.sh` before builds. `nproc` printed **5**, with cgroup quota `400000 100000`.

`bash cloud/setup.sh > /tmp/lint-helpers-setup.log 2>&1` succeeded. Its timing lines, copied to `evidence/setup.log`:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (200s)
setup: done in 200s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/evidence/helpers-final.log 2>&1`: **PASS, 93.006s**. 22,347 cases produced 22,412 identical output lines across Go cohere, Node source and sanitized native. The difference is message values carrying newlines, not extra test cases. There are 15,139 schema cases, 6,991 strict target cases, 186 message cases and 31 grammar cases. 5,923 schema cases carry pinned upstream ESLint verdicts, which the Go oracle independently checks. Ten message invariant failures match Go panic text and exit 70 on Node and native. Explicit regex/custom/Unicode-fold refusals and recovery after a refused call pass.
- `go test ./stage1/cohere/lint/helpers -run '^TestKnownGapsAreExplicit$' -count=1 -v > .../evidence/gaps.log 2>&1`: **PASS, 3.418s**.
- `python3 stage1/cohere/lint/helpers/testdata/regenerate.py > .../evidence/regeneration.log 2>&1`: **22,347 cases; 94 schemas; 54 message consumers**. SHA-256 comparison of catalog, descriptors, compressed cases and coverage: **PASS, identical bytes**, in `evidence/reproducibility.log`.
- `go vet ./... > .../evidence/vet.log 2>&1`: **exit 0, empty log**.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > .../evidence/oracle.log 2>&1`: **PASS, 12.341s**, all six input fixtures, zero cache hits and six probe misses.
- Final isolated mutant rerun: `go test ./stage1/cohere/lint/helpers -run '^TestHelperMutants$' -count=1 -v -timeout=10m > .../evidence/mutants.log 2>&1`. **PASS, 47.340s**. First differing output lines: JSON 15157, schema 15166, strict target 15178, policy 22166.

The full repository gate was not run; this worker used the bounded package and filtered-oracle gate. The touched helper package, repository-wide vet and a filtered uncached external oracle were run. Native builds use ASan/UBSan and Linux's default leak checking; every successful case run must exit 0 with no stderr.

## Mutants

Every final mutant compiles and runs successfully before its wrong output is compared to Go. A compile failure, panic, sanitizer finding or stderr would fail the harness and is not credited as a semantic mutant caught.

| File | Mutation | Independent witness |
|---|---|---|
| options_json.ts | permit a raw control character inside a JSON string | Go rejects the raw newline; the mutant accepts it |
| option_schema.ts | oneOf accepts more than one matching branch | overlapping number/minimum branches: Go refuses `[1]`, mutant accepts |
| strict_options.ts | skip an unknown or wrong-case field | Go refuses an undeclared field; mutant accepts |
| policy_message.ts | omit value interpolation | Go renders the supplied value; mutant leaves `{{name}}` in a finding |

All four were caught both in the complete package run and in the final isolated rerun. No production file is mutated: variants are built from temporary copies. The baseline is the same source on Node and native against Go's helper behavior on consumer schemas, actual Go target types and actual message templates.

## Findings while building

Observations: the first JSON oracle incorrectly used float-decoding `json.Unmarshal`; `1e400` exposed that mismatch with the schema layer's number-preserving decoder. It now uses `UseNumber` and an EOF check. A regexp field in id-length exposed Go's `TextUnmarshaler`; those types are now marked unsupported alongside custom JSON decoding. Native compilation refused a conditional panic expression, stale narrowed mutation state, and an instance method called before every constructor field was initialized. The source was rewritten within the supported subset; no compiler file was edited.

A corpus that repeated large schema strings exhausted Node's default 4 GB heap, and reparsing each descriptor made native validation unnecessarily slow. The corpus now deduplicates 168 definitions and caches schema/target instances. The superseded slow run was stopped; it is not counted as a pass. The constructor-failure log in `evidence/helpers.log` is an early failed attempt, not final evidence. Final compilation, comparisons and mutants passed after those changes.

Inference: these option and policy prerequisites give the largest first two marginal gains in the frozen inventory, 30 and 16. The ledger is a dependency calculation, not an observation of 46 rules running on native.

## Not covered

- The remaining helper packages, all checker/binding questions, new rule implementations and integration into `Settings`/the existing linter.
- Arbitrary configurations or every consumer rule's whole findings/fix corpus. Samples and synthetic cases are bounded; only helper behavior is compared here.
- Full Go decoded struct mutation/default/merge behavior. The strict helper validates target acceptance; per-rule typed construction and custom decoding remain local.
- Regex schema keywords, boundaries extension/descriptor processing, custom JSON/text marshalers, and Unicode folding for untagged keys. These are explicit blockers. The reader limits nesting to 512, which is below Go's maximum; callers must report its error.
- Exact Go schema diagnostic prose, arbitrary malformed catalog loading, global message-claim concurrency/swap APIs, and arbitrary control/non-ASCII identifiers in panic text.
- Refreshing the inventory's stage1 branch observations. Readiness uses its documented frozen 198-rule cohort, not new work other branches may have landed since.
