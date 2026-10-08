# ecmascript/literal package claim

Branch: lint-helpers/ecmascript-literal. Base: origin/area/stage1-lint.
Triage d7ab0bc4 rank 22; no retained port, four missing helpers: CookedToRaw,
cookedBytesProducedBy, escapeWidthInStringLiteral, producesNoCookedBytes.
All 1,516 origin refs checked for ports and all helper branch claims inspected.
Earlier eligible packages are reserved, excluded tonight, or require runtime regex.
Yielded rules/next to earlier remote dcaf0a30 (local claim never pushed).
No literal package reservation or implementation found; regexpattern's dependency
mention is not a literal reservation. This package scans literals, without regex compilation.

Consumer: no-regex-spaces. Forecast: zero alone; one additional with preceding
complete packages (131 cumulative). Its other dependency ecmascript/regexpattern
is reserved: use its completed shared port if available, otherwise stop dependent
rule work explicitly. Counts are conditional, not findings parity claims.

Claim before code; fetch again and yield to earlier competing claim. Build fresh,
one helper per file; actual Go consumer captures compared across Node, emitted
JavaScript and sanitized native; one caught Node/native mutant per helper.
Prove the consuming rule if its shared prerequisites are available. Run helpers
and lint with every input set before the finished-unit implementation push.


## Completed package

All four helpers are complete, built fresh on this branch (no donor literal port):
- CookedToRaw: ../literal/cooked_to_raw.a
- escapeWidthInStringLiteral: ../literal/escape_width_in_string_literal.a
- cookedBytesProducedBy: ../literal/cooked_bytes_produced_by.a
- producesNoCookedBytes: ../literal/produces_no_cooked_bytes.a

Merged verified dependencies ecmascript/regexsyntax bca49a5a and
 ecmascript/regexpattern c55c6335. Shared hex checks and UTF-8 byte projection are
used directly; no private duplicate or runtime regular-expression compilation.
The former dependency stops are resolved. CookedToRaw exposes a tagged Mapping
so Go nil remains distinct from the valid empty cooked string's [0].

Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.
Live unchanged-Go capture covers every consuming-rule upstream case:
no-regex-spaces: 153 cases, 535 helper calls (459 produced, 19 width,
56 mapping, 1 tail); no-misleading-character-class: 261 cases, 1,449 calls
(1,182 produced, 181 width, 82 mapping, 4 tail). There are 189 controls.
All 2,173 captured calls match fresh Go on source Node, emitted JavaScript,
and ASan/UBSan native (9,099 identical output bytes). One compiling output mutant per helper is caught on
all three runtimes. See ../literal/helpers_test.go and ../literal/testdata/coverage.json;
regenerate fixtures and leaf-case counts with ../literal/testdata/capture.py.

Rule proof: ../../rules/no-regex-spaces/. Every one of its 153 upstream cases
runs under the shared Go capture; identical source/rule/options inputs deduplicate
to 95 rows. Selected-rule and full-dispatch findings, spans and fixes agree with Go
for those inputs and the four-finding owned witness. The first-run mutant is
caught on source Node, emitted JavaScript and native, including a sanitized native
canary. See ../../rules/no-regex-spaces/proof_test.go. No helper or rule is stopped.
This finishes the package's single frozen-cohort consumer, no-regex-spaces.

## Finished-unit package gates

Complete helpers tree passed with no skips. Complete lint package passed with
ADAMIC_TYPESCRIPT_SOURCE at pinned v6.0.3 (050880ce59e30b356b686bd3144efe24f875ebc8),
ADAMIC_LINT_BENCH=1, profile generation and profile snapshots enabled.
All 86 registered rule mutants passed on Node, emitted JavaScript and native.
Full upstream parity: 4,987 unique combinations, 14,075,670 identical bytes.
Compiler and stage1 parity: 963 files, 30,602,508 identical bytes.
Profiles, shards and witnesses passed. The existing TestCheckerBridgeRefusalPending
skip remains for the documented compiler gap; no input-dependent test skipped.
The owned rule proof also checks its mutant with ASan/UBSan native.
