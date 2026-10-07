# String-keyed records

Pinned source: TypeScript 6.0.3, commit
050880ce59e30b356b686bd3144efe24f875ebc8. Adamic base:
ef3d907ecdc4c771b016f7d9c52372def057a340.

This bucket answers the census reason `an index signature`: 11 declarations in
7 files, 5 explicit Record<string, T> references in 4 files, and 68 element
accesses on receivers with a string index type. Those 68 are accesses, not reads
alone. The ruled representation is preserved in ../../../docs/index-signatures.md.

Each upstream function body is preserved. MapLike is copied from corePublic.ts.
Support declarations are reduced to the exercised types. The four option fixtures
keep the complete parseOptionValue function, with small dependency stand-ins:
validateJsonOptionValue returns the supplied primitive; unexercised diagnostic,
list and custom-option paths have simplified helpers. They do not test validation
or diagnostics. No adapter or compiler implementation is changed.

| Fixture | Upstream use and driver coverage |
| --- | --- |
| 01_has_property | hasProperty, own membership versus inherited prototype names |
| 02_get_property | getProperty, own toString, missing key and guarded constructor |
| 03_own_keys | getOwnKeys, for-in, post-creation writes and present-undefined |
| 04_own_values | getOwnValues, guarded for-in indexed reads |
| 05_integer_order | getOwnKeys across literal and later keys, canonical indices, 01 and 2^32-1 |
| 06_delete_readd | commandLineParser's delete statement, followed by re-add and enumeration |
| 07_optional_view | getProperty plus a design boundary driver: Record viewed as an all-optional named alias, with writes visible through both views |
| 08_option_boolean | complete parseOptionValue, strict written dynamically and read by name |
| 09_option_undefined | complete parseOptionValue, null argument materializes an own undefined field |
| 10_option_wrong_type | complete parseOptionValue, deliberately inconsistent metadata writes string into strict |
| 11_option_string | complete parseOptionValue, locale string written dynamically and read by name |
| 12_nested_entries | complete nested getScriptTargetFeatures literal table, Object.entries into nested Maps |
| 13_strict_option | getStrictOptionValue, finite string key and strict-default fallback |
| 14_group_by | groupBy implementation, Record<string, T[]>, read/write through ??= |
| 15_inherited_read | same groupBy, deliberate constructor boundary input |
| 16_built_strings | getProperty, dictionary keys and string values built at runtime for future sanitizer work |

The optional alias and wrong metadata drivers are explicitly design probes,
not copied upstream call sites. The delete fixture copies the actual delete
statement, not its containing wildcard-reduction function. The memoize stand-in
in 12 returns the callback: the driver calls it once, so memoization is irrelevant.

## Observations and reproduction

Run from the repository root, after sourcing the setup script's env.sh:

```sh
python3 stage3/fixtures/records/observe.py > /tmp/records-observe.log 2>&1
python3 stage3/fixtures/records/verify.py > /tmp/records-verify.log 2>&1
```

observe.py invokes exactly Node's source oracle and `go run ./cmd/adamic build`
with an explicit output path for every fixture. It records complete command
output, including go run's exit-status line, in status.json. It runs any successful
native build and compares stdout, stderr and exit status. Scratch logs and binaries
are under /tmp/records-observations. No Node runner changes were needed.

On current main, 9 fixtures are Refused and 7 are Checker. None compiles.
Index signatures still say `Adamic 0.1 refuses an index signature`, despite the
new ruled NotYet policy. Checker blockers include upstream indexed-array density
in groupBy/getOwnValues and parseInt(args[i]) in parseOptionValue. They remain
visible instead of rewriting upstream statements to bypass them. The nested
feature table is refused by mutable invariance. This is a baseline, not proof
that the record runtime is implemented. No silent miscompile was observed;
there was no successful native build.

verify.py checks exact Node observations and status coverage/schema. It runs two
source mutants in scratch files: removing getProperty's own-property guard, and
reversing getOwnKeys enumeration. Both finish with exit 0 and are caught by stdout.
These prove the fixture output comparisons can fail, not that a native dictionary
check works. The requested runtime mutants (pure insertion order, optional miss,
inherited panic, named-slot conversion, early value free) were not run because
these fixtures do not reach a native executable on this base. No ASan result is
claimed. The five ruled implementation mutants remain acceptance work for the
compiler feature owners.

10 and 15 require loud-check treatment once the feature lands. Node deliberately
accepts 10 and prints `after write yes`; Adamic must instead exit 70 at the write,
print only `before write`, and name strict, boolean and string in its panic.
Node's 15 prints `before read`, then the source oracle reports
`TypeError: array.push is not a function` with exit 70. Adamic must panic at the
inherited read itself, naming constructor, with no `after read`. status.json
retains the actual source-oracle results, not fabricated future native results.
The ordinary byte-comparison runner must be extended with the established loud
check convention before comparing these intentional safety differences. That
shared harness is outside this unit's territory and was not edited.

## Complete inherited-key census

[INHERITED_KEYS.md](INHERITED_KEYS.md) contains the compiler-wide result and
locations. The complete pass covers every original compiler file, unrestricted
and any receivers, finite mapped shapes, all helper calls, all actual for-in
bodies, and every non-literal in test. It matches all 68 original census spans.
The initial inherited-key-ledger.json is retained as a historical scoped snapshot;
inherited-key-global.json and inherited-key-provenance.json supersede its counts
and provisional classifications.

Global requested-operation union: **1,041 sites**, consisting of **962 fixed
key sets/domains excluding prototype names, 36 user input, and 43 unknown**.
Of those, **145** have string-capable keys: **66 fixed, 36 user input, 43 unknown**.
The remaining **896** have non-string key domains. All helper calls are included,
including 15 literal-key hasProperty calls; omitting those gives 1,026 fully
dynamic-key sites overall and 130 with string-capable keys. There are 36
hasProperty calls on 32 lines, zero getProperty calls, 27 actual for-in loops
with 57 overlapping body sites, and zero non-literal in tests. The 10 literal
in tests are retained as excluded audit rows. Direct hasOwnProperty.call guards
in the helper definitions and loop bodies are included too (12 sites).

User input has disjoint sources: 11 tsconfig only, 13 package.json only,
3 command line only, 5 tsconfig or command line, 2 tsconfig or package.json
(typesVersions paths), and 2 source module specifiers. This sums to 36. Input origins are recorded per location;
own-property guards do not change a user key into a fixed key.

Reproduce in a scratch checkout of TypeScript v6.0.3 at the pinned commit:
install typescript@6.0.3, @types/node@25.3.3 and
@types/source-map-support@0.5.10 in a scratch node_modules reachable from that
checkout. Generate diagnosticInformationMap.generated.ts with upstream's
scripts/processDiagnosticMessages.mjs, then run from Adamic's root:

```sh
export CENSUS_TYPESCRIPT=/tmp/records-census-api/node_modules/typescript/lib/typescript.js
node stage3/fixtures/records/count-inherited-keys.cjs /tmp/records-typescript /tmp/records-global-count > /tmp/records-global-enumerate.log 2>&1
python3 stage3/fixtures/records/summarize-inherited-keys.py /tmp/records-global-count/raw.json > /tmp/records-global-summary.log 2>&1
node stage3/fixtures/records/audit-inherited-keys.cjs /tmp/records-typescript > /tmp/records-global-audit.log 2>&1
```

Stock tsc reports zero diagnostics with these declarations. The audit separately
parses every original source file and uses TypeScript's assignment-target API
for read/write roles, rather than the census scanner's role function. It checks
pinned source hashes against the original census, every source operation, exact
census span identity, loop contexts reconstructed from parent pointers, global
class totals, and source origins. Seven ledger mutants are caught: omitted
element read, omitted helper call, omitted for-in context, omitted in enumeration,
assignment target counted as a read, invented fixed classification, and erased
user-input origin. These are count/audit mutants, not runtime compiler mutants.
Logs are committed in logs/global-*.log.

## Limits

The standalone native fixtures do not cover JSON.parse/fromEntries, null-prototype construction, __proto__ writes,
prototype mutation, record spread/freeze, number index signatures, or cycle
ownership probes. JSON/unknown and runtime mutation need their feature owners;
this bucket prioritizes the requested original compiler operations. No compiler
runtime, adaptation, other bucket or fixtures_test.go was changed. The full
uncached integration gate was not run for fixture-only changes; all 19 individual
build attempts and five fixture mutants were run.

Setup: Go go1.27.1, clang 20.1.8 with its sanitizer probe, Node v24.19.0.
Go/clang/Node/submodules ready at 0s; build cache warm and total at 168s.
nproc 5; cgroup cpu.max 400000 100000. Logs are in logs/.

## Missing inherited keys: ownership review

INHERITED_KEYS.md now gives a per-operation verdict for all 36 user-input and
43 unknown sites. own-guard-verdicts.json is the machine-readable review, kept
separate from the original key-provenance count. The report generator preserves
the ownership section through render-guards.py.

Fixtures 17 and 18 retain compareDataObjects and exercise the two read-first
src[e] sites with missing prototype-named keys and matching own-key controls.
Fixture 19 catches its incorrect equality result on distinct empty records.
The helper/API candidate is separate from the demonstrated tsc CLI candidate.

input-fixtures contains 30 complete JSON/CLI cases run on unmodified stock tsc
6.0.3, with per-case exact stdout, stderr, exit, expected diagnostic and behavior.
probe-inputs.py copies authored .a source to temporary .ts filenames, then invokes
the installed stock tsc.js. Set STOCK_TYPESCRIPT to the typescript@6.0.3 package
directory; --observe refreshes observations.json. probe-guards.cjs uses that same
stock package to check the helper/API result, config own-key loss, host-object
limitation and all 79 ownership-row IDs.

Run both probes and verify.py with output redirected to logs. The new mutants
replace each missing key with an own key, add the missing equality ownership
check, replace __proto__ in a real path mapping with constructor, and omit one
ownership-review row. The compiler/runtime was not changed. Three unknown reads
have no established real CLI/config counterexample, and no full watch-mode
wrong result from compareDataObjects is claimed. UPSTREAM_CANDIDATES.md gives
both scopes and drafted issues; the shared upstream ledger was not edited.
