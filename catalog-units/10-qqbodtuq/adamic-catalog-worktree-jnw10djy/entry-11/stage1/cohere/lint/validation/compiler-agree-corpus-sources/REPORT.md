# Compiler agreement sources-only corpus

Branch `lint-rules/compiler-agree-corpus-sources`, based directly on the gated
`4678c7302cae4ab51e2378156e228e9a771e73b4`. The original guard branch is unchanged.

`TestCompilerAndStage1Agree` excludes a tracked stage1 input if any parent
directory is exactly `testdata`, `validation` or `evidence`. The test logs the
excluded count. Directory prefixes and basenames are not exclusions.
The existing explicit generated production registry and pinned compiler sources
remain covered.

The inventory is 790 tracked stage1 inputs: 243 fixtures excluded and 547 sources
retained. With 77 compiler sources and one generated registry, agreement covers
625 inputs. All current exclusions are under `testdata`; the other two folder
names are covered by the regression fixture. [excluded-paths.txt](excluded-paths.txt)
names every excluded file; [inventory.json](inventory.json) records counts.

The regression fixture commits 18 inputs in a temporary Git repository. Its
seven named-folder fixtures contain `new Promise(@dec async () => {})`; all 11
source inputs, including lookalike folder names and ordinary `fixtures`/`tests`
folders, must remain. The test also fixes the allowed folder-name list explicitly,
so extending that list fails. [fixtures.txt](fixtures.txt): pass, zero skips.

Three mutants were caught by the regression test, with the expected assertion
rather than a build failure: retaining fixtures, adding `tests` to the folder
policy, and excluding the `testdata` prefix. [mutants-summary.json](mutants-summary.json)
records durations; individual `.txt` logs and `.patch` files retain the failures.
`run-mutants.py` reproduces the Go-overlay checks without changing the checkout.
[vet.txt](vet.txt): vet passed for lint, its changed helpers, and internal/testguard.
`tested-code.json` records the tested code hashes and commands.

Compiler agreement passed in **329.17s**, with one pass, zero failures and
zero skips. All four backends returned identical **29,860,490 bytes** for 625
inputs. Longest shard: sanitized native 0/5,
**196.786s**, containing `checker.ts` alone.
[agreement-summary.json](agreement-summary.json) and [agreement.jsonl](agreement.jsonl)
retain timings and root/exclusion counts. No full package
rerun was needed for this narrow follow-up: completed normal and every-core-loaded
package proofs for the base are in [../compiler-agree-guard-main/REPORT.md](../compiler-agree-guard-main/REPORT.md).
