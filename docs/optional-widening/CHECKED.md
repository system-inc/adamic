Built checked optional reads on Lane 1's shared transitive view machinery, retaining reduced-source and never traps.
Commits: merge 3ececcf13a077a1760303d527a001bc293bb1855 includes Lane 1 609ed39; read checkpoint is the commit containing this report.
Validation: lower 29.673s, fixtures 16.474s, whole uncached Node oracle 291.873s; restored focused oracle 12.885s; setup 48.638s, nproc 5 (quota 4).
Mutants: erase read checks, conflate null with undefined, and skip finite-literal validation all fail semantic assertions with valid C.
Not covered at this checkpoint: checked writes, whole adapted 391-site measurement, class expressions' existing lowering boundary, callable and dictionary view contracts.

## Checked reads, October 7

The optional-widening relation pass now installs `(*lowering).view(node, value, target)` instead of refusing a supported structural widening. The same shared field readiness and representation checks cover object/interface aliases, argument/return boundaries and transitive fields. Optional absence produces undefined; a present incompatible value exits 70 at the read. The six developer widening programs compile and stop with their complete diagnostic pinned, rather than emitting the wrong value. Their original Node observations remain independent and unchanged. Function-typed developer programs are skipped for codex/refusal-pass-rulings.

A representative diagnostic is `adamic: panic: field read failed: wider.y is not a number | undefined; expected number | undefined, found boolean`. Each program is checked in sanitized native, release native and generated JavaScript. Added probes cover absent and present optional numbers and booleans, strings, nested optional objects, optional chaining, explicit undefined, null and finite number literals. Literal null has a distinct representation tag; dynamic nullable hidden fields are conservatively fenced until a complete null certificate exists. Missing required fields remain traps, even when another program installs an optional check for the same name.

The class descendant rule is now a read-erasure condition. It relies on the closed world: every class in every loaded root/module, including transitive subclasses, is considered. Separate compilation requires revisiting this rule. Incompatible direct, grandchild and generic descendants retain the read check and trap, naming the read and expected/found types. Class expressions are considered by the proof, but executing the class-expression fixture still reaches the pre-existing ClassExpression NotYet boundary; it is not claimed as compiled.

Stage 3 fixture records were rechecked under TypeScript's condition. Six assertion fixtures now compile following the Lane 1 merge and agree with their unchanged Node observations. The two original optional-refusal object sites still do not compile: reference spreads return to the single-spread boundary, and the optional host method reaches the callable-view boundary. Neither was formerly Compiles. Only stage0 record values changed; Node record bytes are preserved. Other records reflect the earlier boundary reached on this merged compiler.

## Commands and evidence

All test output went to files. The complete final gate was `ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./stage3/fixtures ./internal/oracle -count=1 -timeout 30m`. Lower and the whole oracle passed; the single stale assertion diagnostic record was then corrected and the complete fixture package rerun successfully. No compiler implementation changed between that gate and the fixture rerun. The final focused oracle command was `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestOptionalCheckedRead|^TestOptionalClassReads|^TestOptionalNullAndLiteralReads$' -count=1`.

The read-marker mutant changes the shared CheckedFields registration to false: all six developer programs then exit 0, caught by the exit-code assertions. The null-tag mutant erases the distinct null tag: the null probe incorrectly continues, caught by exit code. The literal mutant skips finite-literal validation: the value 1 is accepted as 2, caught by exit code. All generated C is valid. Sources are restored in finally, and the focused oracle passes afterward. The runner for shared read erasure is evidence/read-mutants.py; additional mutant logs identify the exact changes.

Setup completed successfully with `/workspace/adamic-tools/env.sh`: Go 0.040s, Node 0.034s, submodules 0.134s, clang 0.395s, Markdown 1.310s, build 48.362s, warm 48.591s, done 48.638s. No module-fetch workaround was needed. The scoped vet and whitespace checks pass. The full repository gate is not claimed.


## Write checkpoint supersedes the read-only status

Checked writes and the final stage 3 measurement are now recorded in [WRITES.md](WRITES.md). This document's opening five lines describe the earlier read push `d3226902`; writes are no longer left pending. The final supported write path checks declared slot contracts at the write and retains conservative proof erasure.
