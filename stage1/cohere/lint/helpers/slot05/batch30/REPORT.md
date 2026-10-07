# Batch 30 handoff

Built colorArm, themeArm and widthArm, one Adamic .a helper per file. Each removes a listed prerequisite for better-tailwindcss/enforce-consistent-class-order, better-tailwindcss/enforce-shorthand-classes and better-tailwindcss/no-unknown-classes. Nine dependency occurrences across three distinct rules; none becomes newly helper-ready. Cumulative slot05: 86 helpers, 419 occurrences, 76 consumers, 51 helper-ready candidates including the common 46. These are frozen dependency calculations, not implemented rules.

Claim efb04ae72 was successfully pushed before any new source was written. The scan read claims on all 20 origin codex/lint-helpers* branches and shared HELPERS.md. Each selected helper tied the highest available concrete fan-out of three. Both origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint db2ecc00447f9ebe8adecb190f71ac222e5db860 were unchanged and already ancestors of the published branch. Prior 83 helpers and compiler/runtime inputs were unchanged; their retained proofs remain applicable. No shared file was authored or main/area branch pushed.

## Observed validation

Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db is asserted by the suite. Its unchanged constructors supply verdicts through test-only exports. All three modes capture 406 distinct original fixture strings from every inventory consumer and retain 424 values after controls. Six parameter scenarios per value produce 2544 cases per helper, 7632 total. Every baseline agrees byte-for-byte across actual Go, source Node, emitted JavaScript on Node and sanitized native with exit zero and empty stderr.

- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch30 -count=1 -v -timeout=20m: PASS, 30.407s, including all 24 native semantic mutants.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v: PASS, 1.023s, seven actual input probes, seven misses and zero cache hits.
- go vet ./...: exit zero, empty output. gofmt -l cmd internal stage1/cohere/lint/helpers/slot05: exit zero, empty output.
- export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh: PASS, 24.455s; nproc=5, quota 400000 100000, 17.6 GB. Go ready 0.023s, Node 0.022s, markdown 0.064s, submodules 0.080s, clang 0.207s, Go build 24.306s, test binaries deferred 24.428s, cache warm 24.430s. Go 1.27.1, clang 20.1.8, Node 24.19.0. source /workspace/adamic-tools/env.sh before test commands.

All test output went directly to named log files. Lossless logs, every generated case and actual Go verdict, coverage file lists, mutant witnesses and identities are in evidence/. The source tree had sufficient disk space; no prior published evidence was deleted.

## Mutants

Eight per helper, 24 total, compiled and ran with exit zero and no stderr before an output mismatch caught them. Full replacements and first independent Go witnesses are in evidence/mutants.json and helpers.log.gz.

All three: invent an InferTypes slice; enable percentage passthrough; erase theme presence; flip the color flag; change the bare-value kind; change the suffix; change the modifier flag. The eighth color mutant copies the supplied slice, caught by Go's alias behavior. The eighth theme and width mutants omit their single-key arrays, caught by field/length checks. Source and emitted baselines are tested; mutants run on native only. No compiler or sanitizer failure is credited as a semantic kill.

## Limits

The full gate, its 17 required external correctness checks, full shared lint suite, whole-rule findings/fixes/suggestions parity, performance and arbitrary invalid presence-flag combinations were not run. No skipped check is claimed green. Resolution and CSS behavior remain separate dependencies. No regex is needed by these constructors.
