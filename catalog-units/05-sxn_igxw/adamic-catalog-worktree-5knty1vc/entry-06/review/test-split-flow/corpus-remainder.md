# Corpus remainder verification

The generated tables now require existing, unique corpus paths and the correctly named function for each path and analysis family. New corpus paths are checked by `TestFlowCorpusRemainder` and `TestFreshCorpusRemainder`, each through the original checking functions in its own subtest. The aggregate remainder uses the existing budget guard: over 30 seconds fails only with `ADAMIC_UNIT_BUDGET=1`, and otherwise logs. A long remainder signals that the table should be regenerated.

Base: test-split-flow e6a6174e merged with main 5ff7d880 (initial verification on 8afb1ee4). Running `python3 review/test-split-flow/generate.py` through gofmt generated 4105 family bindings without changing the existing tables. Flow has 826 generated programs and fresh has 801. Both remainders start at zero.

Commands and observations:

- `go test ./internal/flow ./internal/fresh -run 'Corpus(UnitsCoverEveryProgram|Remainder)$' -timeout 90s -count=1 -v`: passes with no remainder programs.
- Add temporary `internal/flow/testdata/corpus_remainder_probe.a`, absent from both tables, with `const value: string = "remainder control"; console.log(value);`: both coverage tests pass and both remainder tests pass, each reporting one program.
- Change the temporary program to `const value: number = "planted remainder failure"; console.log(value);`: both coverage tests still pass, and each remainder subtest fails on the planted load/type error. This proves new programs reach the actual checking functions.
- Temporarily rename `internal/flow/testdata/joins.a` outside the corpus glob: both coverage tests fail naming the missing generated path.
- All temporary inputs were restored. No new oracle fixture remains, so no counts row changes are required.
- `go test ./internal/flow ./internal/fresh -timeout 90s -count=1`: both packages exceeded the aggregate 90-second timeout, without a reported assertion failure before timeout. This is not a complete package pass.
- `go test ./internal/flow ./internal/fresh -run 'Corpus|Program_testdata_|Writes____flow_testdata_' -timeout 90s -count=1 -v`: passes both packages, including generated joins and mutations units. Repeated successfully after merging final main 5ff7d880.
- `go vet ./internal/flow ./internal/fresh`: passes.

Logs are in `/workspace/corpus-final-controls.log`, `/workspace/corpus-added-control.log`, `/workspace/corpus-remainder-mutant.log`, `/workspace/corpus-missing-mutant.log`, `/workspace/corpus-packages.log`, and `/workspace/corpus-vet.log`.

Setup completed in 49.521 seconds; nproc is 5. Readiness timings: Node 0.143s, Go 0.149s, submodules 0.293s, markdown dependencies 0.398s, clang 0.905s, build 47.430s, cache 49.377s. No compiler or runtime files changed. This serves roadmap step 38.
