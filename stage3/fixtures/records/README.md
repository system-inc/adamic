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

## Inherited-key provenance: scoped result, not the requested global count

inherited-key-ledger.json preserves all 68 census rows and adds classifications
and evidence. Locations distinguish accesses on the same line by column.
There are 22 assignment/delete-only sites, one literal dot-key read, and
**45 dynamic reads**: **11 fixed-set keys excluding prototype names, 24 keys
from user input, and 10 unknown**. Compound ??= and ||= count as reads.
Own-key guards are kept distinct from key provenance: user keys can themselves
name prototype members while being present own properties, which is safe under
the ruling. Pattern-only keys, dot-prefixed exports and #-prefixed imports have
additional exclusion evidence in the ledger. Fixed-set reasoning assumes the
unmodified pinned internal option tables and no external prototype mutation.

This is not every dynamic-key read in src/compiler. The census explicitly covers
receivers with string index types; it omits any receivers, some generic object
operations, numeric accesses and finite mapped shapes. A complete independent
AST/type traversal plus caller provenance review was not completed here. The
45 subtotal must not be published as @system_adamic's requested global count,
and the 34 user/unknown subtotal is not a count of unguarded inherited misses.

## Limits

No JSON.parse/fromEntries, null-prototype construction, __proto__ writes,
prototype mutation, record spread/freeze, number index signatures, or cycle
ownership probes. JSON/unknown and runtime mutation need their feature owners;
this bucket prioritizes the requested original compiler operations. No compiler
runtime, adaptation, other bucket or fixtures_test.go was changed. The full
uncached integration gate was not run for fixture-only changes; all 16 individual
build attempts and both fixture mutants were run.

Setup: Go go1.27.1, clang 20.1.8 with its sanitizer probe, Node v24.19.0.
Go/clang/Node/submodules ready at 0s; build cache warm and total at 168s.
nproc 5; cgroup cpu.max 400000 100000. Logs are in logs/.
