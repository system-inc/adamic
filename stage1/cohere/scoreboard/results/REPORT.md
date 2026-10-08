# Step 43: all 235,377 tracked files at the 23 public pins

Both snapshots are **dependency-uninstalled**. This is the identified public subset of Kirk’s quiet hundred; it does not include the other Mac-only snapshots. The port is the source Node execution, not a native Adamic lint build.

**235,377 files; 201,196 script files; 1,195,182,063 input bytes.** All eight requested extensions are included. 51 tracked symlinks use their Git blob bytes, preserving the original paths.

**99,626,201 cells: 17,175,641 agree, 11,476 diverge, 82,439,084 blocked.** Agreement among successfully compared cells: 99.933229%. Blocked cells are excluded from that percentage and never counted as matching.

Scope: 93 registered port rules, 399 unported registry rules, full all-fixes, a separately labeled 90-listener syntax-fixes comparison, and one formatter per input. Full all-fixes remains blocked when typed rules require checker facts. Syntax-fixes certifies only the measured syntax listener scope.

## Agreement per family

| Family | Checks | Agree | Diverge | Blocked |
| --- | ---: | ---: | ---: | ---: |
| @eslint-community | 201,196 | 0 | 0 | 201,196 |
| adamic | 1,408,372 | 188,441 | 16 | 1,219,915 |
| base | 3,219,136 | 753,828 | 0 | 2,465,308 |
| better-tailwindcss | 2,414,352 | 0 | 0 | 2,414,352 |
| boundaries | 201,196 | 0 | 0 | 201,196 |
| core | 33,800,928 | 9,987,873 | 348 | 23,812,707 |
| fixes | 402,392 | 188,325 | 132 | 213,935 |
| format | 235,377 | 26,990 | 10,540 | 197,847 |
| next | 4,426,312 | 376,914 | 0 | 4,049,398 |
| nexus | 10,462,192 | 1,884,236 | 334 | 8,577,622 |
| react | 12,876,544 | 1,696,104 | 9 | 11,180,431 |
| react-hooks | 3,420,332 | 0 | 0 | 3,420,332 |
| structure | 6,639,468 | 188,457 | 0 | 6,451,011 |
| typescript | 19,918,404 | 1,884,473 | 97 | 18,033,834 |

Every rule’s numbers, findings and successful-cell percentage are in `per-rule.csv`; every file’s states, hashes, finding counts, stdout digests, exit codes and timings are in `files-*.jsonl.gz`. Rule order and the state alphabet are in `summary.json` and SCOUT.md.

## Blocked cells

| Cause | Cells |
| --- | ---: |
| checker fact | 753,828 |
| execution fault | 6,126 |
| formatter refusal | 196,905 |
| parser | 1,205,021 |
| unported rule | 80,277,204 |

Execution faults are kept separately where the receipt cannot establish one of the requested semantic causes. Unported rules have an additional Go census status histogram; those statuses overlap the primary unported block and must not be added to the primary total.

## Divergences and responsibility

All **11,476 divergence cells** have same-path replayed, source-subsequence reductions, complete reduced Go/Node answers and checked implementation references in `divergences-minimized.jsonl.gz`. There are **325 distinct rule/signature/input witnesses** in `witnesses.json`. Reduction is a single-rune deletion fixed point with Go parse validity held for scripts; global shortestness is not asserted.

The inspected causal sites distinguish comment/BOM loss, blank lines, hashbang comment guards, import phases, constructor/static-block parsing and missing JSDoc attachment. Every cell retains actual rule emission/fix or formatter entry sites; the witness catalog adds the inspected causal sites. The final paired tree checks are retained in `additional-trees.json`.

Cases with emission-site attribution only: {}.

## Unported rules ranked by measured findings

| Rank | Rule | Go findings | Finding files | Census |
| ---: | --- | ---: | ---: | --- |
| 1 | one-var | 921,962 | 89,282 | incomplete |
| 2 | id-length | 581,633 | 62,755 | incomplete |
| 3 | no-inline-comments | 181,067 | 18,721 | incomplete |
| 4 | structure/consistency-require-organized-imports | 111,605 | 111,605 | incomplete |
| 5 | strict | 73,910 | 15,978 | incomplete |
| 6 | structure/import-require-react-namespace | 65,151 | 11,409 | incomplete |
| 7 | no-empty-function | 58,537 | 15,901 | incomplete |
| 8 | @typescript-eslint/no-non-null-assertion | 49,243 | 8,173 | incomplete |
| 9 | class-methods-use-this | 43,188 | 12,804 | incomplete |
| 10 | dot-notation | 28,775 | 4,220 | incomplete |
| 11 | @typescript-eslint/prefer-as-const | 24,214 | 111 | incomplete |
| 12 | sort-vars | 21,982 | 2,412 | incomplete |
| 13 | arrow-body-style | 17,743 | 8,748 | incomplete |
| 14 | adamic/single-spread | 17,085 | 8,134 | incomplete |
| 15 | @typescript-eslint/array-type | 17,076 | 6,032 | incomplete |
| 16 | @typescript-eslint/no-empty-object-type | 16,871 | 5,616 | incomplete |
| 17 | structure/react-component-no-destructuring | 15,778 | 11,242 | incomplete |
| 18 | base/consistency-no-bare-throw | 14,262 | 5,748 | incomplete |
| 19 | max-lines | 13,414 | 13,414 | incomplete |
| 20 | @typescript-eslint/prefer-enum-initializers | 12,986 | 1,680 | incomplete |

The complete 399-rule ranking is `unported-ranked.csv`. An incomplete census is a lower bound: missing checker programs, required options, option decoder failures and parse diagnostics remain counted and named. This measures registry defaults, not installed project policies.

## Receipt integrity

All 235,377 oracle records passed the byte-transport audit; 0 files needed corrected answers. JSON’s only lossy string substitution is U+FFFD: explicit raw-byte fields and old strings without that sentinel are lossless; every ambiguous old string is independently replayed.

Source Node and the real Go oracle ran every applicable input. Persistent host workers reuse transport, not language implementations. CLI-versus-worker fixtures, an independent isolated-rule census, the live message mutant, invalid-byte mutant and stderr collector regression pass in the package validation receipt.
