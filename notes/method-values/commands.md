# Commands and observations

Commands ran from `/workspace/adamic`, except the cohere build, which ran from `/workspace/adamic/cohere`. Toolchain shells sourced `/workspace/adamic-tools/env.sh`.

## Setup

Ran `bash cloud/setup.sh` twice. The first invocation overlapped checkout and its warm compilation saw the method adapter file before it was available; it failed with undefined lowerer methods. The second ran on the stable checkout and succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (100s)
setup: done in 100s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Go 1.27.1, clang 20.1.8, Node 24.19.0.

## Repository inspection

Read `CLAUDE.md`, `README.md`, `docs/0.1.md`, `docs/memory.md`, the setup script, the branch diff and commit messages, changed lowerer files and tests, the oracle harness and counts harness, and the branch survey report and JSON artifacts. Used `rg` to search existing oracle programs for prototype values and map callbacks.

```
git fetch origin codex/method-values:refs/remotes/origin/codex/method-values
git fetch origin main
git diff origin/main...origin/codex/method-values
git log --oneline origin/main..origin/codex/method-values
git switch -c coverage/method-values origin/codex/method-values
```

The initial diff before refreshing main included unrelated integration history. After refresh it comprised only `57bf6dc` and `3fb24bc`.

## Feature tests

Test output went to files under `/tmp/method-values-coverage/`; no running test was piped to head or tail.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_coverage_' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

Final scoped oracle: PASS, 40 subtests, 7.535s, native hits 0 / misses 118, Node hits 0 / misses 80. Counts: PASS, 16.565s, exactly 40 rows added.

An initial `-run 'TestNativeAgreesWithNode/.*method_coverage_'` selected no subtests because the fixture names contain slashes. Corrected the selector. The first real run caught the optional-boolean array refusal; moved that probe to notes and reran all retained programs successfully.

Python subprocess loops ran the following for each candidate, captured stdout/stderr/exit separately, and ran the output binary only if compilation succeeded:

```
node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute-file>
go run ./cmd/adamic build <file> -o /tmp/method-values-coverage/<stem>
/tmp/method-values-coverage/<stem>
```

The first loop covered 42 candidates. Corrected ordinary typing errors and kept unsupported probes outside testdata. The second covered 41 candidates: 40 agreed and the optional-boolean array was rejected. A third loop ran Node and the CLI build for each of the 36 final notes probes, using `-o /tmp/method-values-coverage/unsupported-probe`; all were rejected and no native execution was possible. Full final observations are in `observations.json`. Invalid interim spellings are documented by the initial `.build` logs and are not counted as additional feature programs.

## One-line mutant

Changed only `tag = "Map"` to `tag = "Set"` at `internal/lower/library_method_values.go:382`, then ran:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_coverage_object_tags.a$' -count=1 -v
```

It failed with stdout differences in native and JavaScript backends, both exit 0, no stderr, and no sanitizer/leak failure. The source on Node printed `[object Map]` on its first line; both mutated backends printed `[object Set]`. All following lines matched. Verified restored source byte equality with `git show HEAD:internal/lower/library_method_values.go` and reran the same command: PASS, 0.554s. The first restoration used an ambiguous replacement and its assertion caught the wrong Set line; corrected that block before the final successful rerun.

## Repository gate

```
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
git diff --check
```

Formatting and vet produced no output and succeeded. Full gate status is recorded below after completion.

## Fixture formatting

```
cd /workspace/adamic/cohere
go build -o /tmp/method-values-coverage/cohere ./command/cohere
```

Prepared exact `.ts` copies of the 40 fixtures, the prelude and compiler settings under `/tmp/method-values-coverage/format`, because the pinned CLI ignores `.a` paths.

On the temporary copies, ran:

```
/tmp/method-values-coverage/cohere --format-only --no-cache *.ts
/tmp/method-values-coverage/cohere --no-fix --no-format --no-cache method_coverage_*.ts
/tmp/method-values-coverage/cohere --format-only --no-cache method_coverage_*.ts
/tmp/method-values-coverage/cohere --no-fix --no-cache method_coverage_*.ts
```

The first check reported intentional unbound-method and no-useless-call uses, unsafe array assignment annotations whose element equality the branch proves, the intentional temporal-dead-zone read, and unnecessary string templates. Added narrowly applicable fixture comments like the existing branch fixture, removed the unnecessary string templates, applied the formatter diff back to the `.a` files, then the final check passed with no findings: `0.2s (276 rules • 40 checked) • 25% Adamic-ready (10 of 40)`. The other 30 use documented oracle exceptions; 100% readiness is not claimed. The warning about 40 of 41 files refers to the prelude, which was not requested for lint.

Reran the same scoped uncached oracle on the final formatted/commented files: PASS, 40 subtests, 23.404s, zero cache hits. Reran the counts command and the Node/build/native CLI loop on all 40 final `.a` files as well. Final statuses follow after completion.

Final counts check: PASS, 62.558s; still exactly 40 added rows, no existing row changed. The final CLI loop rebuilt and ran all 40 files and asserted equality of stdout, stderr and exit status against Node; all passed. Final observations replaced the earlier candidate observations in `observations.json`.

One additional probe takes `trim` directly from a string instance and uses `.call(text)` later. Ran the same Node and CLI build commands for `notes/method-values/unsupported/string_instance_method.a`: Node printed `x\n`; Adamic refused the detached instance method. The notes therefore contain 37 rejected programs total. All 40 retained programs use the branch's admitted intrinsic/prototype forms.

## Final additional String overload coverage

Sound callable annotations let `String.prototype.split`, `replace` and `replaceAll` select their string overload for delayed call/apply. Their unannotated notes probes still fail checking, but the annotated forms need no cast and do compile. Added `method_coverage_string_overloads.a`, including split of a supplementary character into UTF-16 code units and literal replacement tokens.

Ran Node, the exact CLI build and the binary first on `/tmp/method-values-coverage/string_overloads.a`, then on the final `internal/oracle/testdata/method_coverage_string_overloads.a`; outputs agree. Repeated the same scoped uncached oracle command with 41 fixtures: PASS, 8.552s, native misses 121 and Node misses 82, all hits zero. Repeated the required counts update: PASS, 25.763s, exactly 41 added rows and no old row changed.

Cohere `--format-only --no-cache method_coverage_string_overloads.ts` followed by `--no-fix --no-cache method_coverage_*.ts` passed with no findings: `0.08s (276 rules • 41 checked) • 24% Adamic-ready (10 of 41)`. Applied the formatted copy to its final `.a` file.

## Completed gate

The full uncached repository gate exited 0. All 30 package rows passed or had no test files. The raw package results are in `gate.log`; native took 562.585s, oracle 447.175s, Unicode properties 1191.907s and stage-one JSON 645.004s. That broad run started with the first 40 fixture registrations; the additional overload program was subsequently included in the final 41-fixture uncached oracle, counts and CLI checks above. Compiler code did not change after restoration.

Ran final `gofmt -l cmd internal`, `go vet ./...` and `git diff --cached --check` before committing.

```
git add internal/oracle/method_values_coverage_test.go internal/oracle/testdata/method_coverage_*.a internal/oracle/counts.md notes/method-values
git commit -m "Add oracle coverage for delayed method values"
git push -u origin coverage/method-values
```
