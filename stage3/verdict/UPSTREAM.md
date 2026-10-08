Expanded the upstream CLI harness from 6,262 to 11,610 input files and 13,453 configurations.
This unit follows 85740c95 on codex/stage3-verdict-harness; one finished-unit push.
A/B group measurements, public-entry proof and full C evidence are in evidence/expansion/.
Every admitted group catches its one-byte diagnostic mutant; three CLI bridge mutants are also caught.
834 inputs remain excluded; 533 admitted inputs have excluded variants, enumerated individually.

# Upstream expansion

Counts distinguish input files from compiler configurations. All 6,262 original inputs remain admitted. The 6,182 original exclusions shrink by 5,348 to 834. Pass counts below count configurations; source files with multiple variants count once in the census. Selection is determined from pinned inputs, references and stock parser metadata, never from the tested binary’s results.

## Original exclusions grouped by reason

| Original reason group | Excluded before | Inputs admitted | Still wholly excluded |
|---|---:|---:|---:|
| option-variants | 1559 | 1477 | 82 |
| filename | 1488 | 1290 | 198 |
| declaration-emit | 1035 | 1008 | 27 |
| syntax-only | 971 | 736 | 235 |
| compiler-directives | 811 | 627 | 184 |
| references | 240 | 154 | 86 |
| global-diagnostics | 78 | 56 | 22 |

The compiler-directive group includes lib selection, config-only options, embedded projects, UTF-16, working directories, symlinks and pretty formatting. Filename cases preserve source text, root selection and relative resolution. Option variants follow the pinned declaration metadata, including wildcard exclusions and alias deduplication. Declaration cases run with their real emit settings and check the CLI exit code.

## Measurements

The group proof measures every added configuration, then changes one diagnostic byte on one new diagnostic configuration in that group. It verifies the mutant preserves stderr and exit and differs in exactly one raw stdout byte in the same working directory. Counts and witnesses are retained in [evidence/expansion/proof.json](evidence/expansion/proof.json).

| Added group | A passes/configurations | B failures/runs |
|---|---:|---:|
| option-variants | 3145/3145 | 1/1 |
| filename | 1305/1305 | 1/1 |
| declaration-emit | 1041/1041 | 1/1 |
| compiler-directives | 752/752 | 1/1 |
| syntax-only | 736/736 | 1/1 |
| references | 154/154 | 1/1 |
| global-diagnostics | 56/56 | 1/1 |
| message-diagnostics | 2/2 | 1/1 |

A coverage is an aggregate of captured runs, not a claim that a single final full invocation was run: 7,191 newly admitted configurations plus all 6,262 original configurations. The retained-input regression initially passed 6,259/6,260; the literal-string-null case then passed with its JSON option bridge. Two initially over-excluded JSDoc cases passed after scope correction. The original failed observation and successful corrections are retained rather than relabeled. All final source, reference and expected-summary hashes are checked against the successful observations.

The public entry-point smoke, before the final two message configurations were added, runs all 301 acceptance projects, tiny and the first 40 upstream configurations. A passes 301/301, 1/1, 40/40. B passes 300/301, 0/1, 39/40; each failure is one stdout byte. C passes 0 in every suite. The separate full C run fails 301/301, 1/1 and 13,453/13,453, with no harness errors. Smoke-deferred configurations are explicitly counted and do not count as passes.

Three real pinned-case implementation mutants prove the delicate CLI bridges: omitting the literal-null JSON bridge fails stdout; forcing declaration exit 2 fails exit; forcing explicit type roots for provided packages fails stdout. The focused tests also kill stdout, stderr, exit, EOF, census and configuration-identity mutants. Source-byte and reference-byte pin mutants both fail before invocation and restore their inputs.

## Remaining wholly excluded inputs

| Reason | Inputs | Why a faithful CLI comparison cannot currently reproduce the pinned reference |
|---|---:|---|
| API diagnostic collection/formatting | 603 | The compiler runner gathers diagnostics that CLI control flow suppresses, discards config-parse errors, or adds API-only consistency assertions. Detailed variant reasons follow. |
| Absolute source references | 172 | The harness mounts virtual absolute roots. Preserving source text would require filesystem root mounts; rewriting import/reference text changes the test. |
| Windows virtual paths | 23 | Drive and separator semantics differ on this Linux host. |
| Absolute package/config paths | 14 | Config and package contents refer to mounted virtual roots; changing their text changes resolution. |
| suppressOutputPathCheck | 13 | Internal host override has no public CLI equivalent. |
| Case-insensitive host | 4 | Linux filesystems do not provide the requested case-insensitive virtual host behavior. |
| Compiler version override | 4 | Three request 5.0 and one 5.5; the tested compiler is pinned to 6.0.3. |
| captureSuggestions | 1 | Suggestion diagnostics are collected through compiler APIs, not the CLI error reporter. |

The 603 wholly excluded API cases split as follows. Combined rows mean different configurations of the same input hit different boundaries; each input is counted once.

| Whole-input diagnostic boundary | Inputs |
|---|---:|
| parser errors suppress other diagnostics on CLI; API collects both | 325 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both | 184 |
| JavaScript syntax errors suppress semantic diagnostics on CLI; API collects both | 50 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both; parser errors suppress other diagnostics on CLI; API collects both | 22 |
| config option diagnostics suppress semantic diagnostics on CLI; API collects both | 8 |
| pre/post emit consistency diagnostics exist only in the API harness | 5 |
| diagnostics outside materialized units require harness-mounted paths or library placeholders | 4 |
| config parse diagnostics are discarded by API compiler runner but reported by CLI | 2 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both; option diagnostics suppress global diagnostics on CLI; API collects both | 1 |
| parser errors suppress other diagnostics on CLI; API collects both; pre/post emit consistency diagnostics exist only in the API harness | 1 |
| option diagnostics suppress global diagnostics on CLI; API collects both | 1 |

A case can have multiple excluded configurations, including a partly admitted case. The 1,268 known excluded configurations have the following disjoint reasons. Host-only cases rejected before variant expansion are counted as inputs, not invented configuration totals. Exact filenames, configurations and reasons remain in selection.json and the run’s configuration-exclusions.json.

| Configuration reason | Count |
|---|---:|
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both | 810 |
| parser errors suppress other diagnostics on CLI; API collects both | 368 |
| JavaScript syntax errors suppress semantic diagnostics on CLI; API collects both | 63 |
| config option diagnostics suppress semantic diagnostics on CLI; API collects both | 8 |
| pre/post emit consistency diagnostics exist only in the API harness | 7 |
| option diagnostics suppress global diagnostics on CLI; API collects both | 6 |
| diagnostics outside materialized units require harness-mounted paths or library placeholders | 4 |
| config parse diagnostics are discarded by API compiler runner but reported by CLI | 2 |

Message-category TS1450 summaries were audited and admitted with their own A/B proof. The four outside-unit diagnostics include mounted/library placeholder paths. Malformed JSON belongs to parser-error gating: CLI reports syntax only while the API baseline also includes semantic diagnostics. CLI syntax-only, global-only, lib-option and declaration-only configurations are admitted; the exclusions above require different diagnostic collection or host semantics, rather than an unimplemented ordinary directive.

## Commands and setup

All test output was redirected to log files. No whole Go packages or full gate were run. No Adamic oracle fixtures were added, so counts.md does not change.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stage3-verdict-unit2-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-adapted
python3 stage3/verdict/prove_groups.py /tmp/verdict-before.json stage3/verdict/selection.json /tmp/stage3-verdict-adapted /tmp/new-group-proof > /tmp/group-proof.log 2>&1
python3 stage3/verdict/prove.py --baseline-limit 40 /tmp/new-entry-proof > /tmp/entry-proof.log 2>&1
stage3/verdict/run.sh --tsc stage3/verdict/standins/empty.sh /tmp/new-full-C > /tmp/full-C.log 2>&1
python3 stage3/verdict/test_verdict.py > /tmp/checks.log 2>&1
python3 stage3/verdict/audit_cli_mutants.py stage3/verdict/selection.json /tmp/stage3-verdict-adapted /tmp/new-cli-mutants > /tmp/cli-mutants.log 2>&1
python3 stage3/verdict/audit_mutants.py /tmp/disposable-pristine-tree > /tmp/pin-mutants.log 2>&1
```

Actual group runs were staged: the first run proved option-variants, filename and declaration-emit; corrected compiler-directives and the remaining three groups ran separately using --group. The final message-category delta ran separately and passed A 2/2 with B 1/1 caught. The first group log retains the three directive mismatches that drove the faithful lookup/JSON fixes. Final successful group reports establish the table above. The pristine checkout reproduces selection.json byte for byte; options and declaration-diagnostic metadata were independently regenerated from pristine sources.

Setup succeeded: Node 0.024s, Go 0.028s, validated markdown dependencies skipped at 0.085s, submodules 0.085s, clang 0.201s, Go build 40.985s, tests deferred 41.085s, cache 41.086s, done 41.115s. nproc is 5; cgroup CPU quota is 4. The native compiler remains unlinked; run the README command as soon as its executable exists.
