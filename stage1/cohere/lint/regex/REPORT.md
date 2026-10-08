Built: fixed RegExp literal table and Go/source-Node/emitted-JS/sanitized-native comparison; rule migrations blocked.
Commits: see git log on codex/lint-regex; base f8013f0baac41ddc340d76f83bddde38536a8f07.
Commands: owned package PASS 10.352s, vet clean; filtered external regexp oracle PASS 11.580s.
Mutant: PaginationInput changed to PaginationInputBROKEN; all three backends ran cleanly then comparison named base/consistency_require_pagination_argument_name.go:27.
Not covered: quiet-hundred manifest, full rule findings/fixes, dynamic option constructors, arbitrary option dialect parity, captures, replacement APIs and full repository gate.

The table has 107 rows: plain 76, (?i) 4, (?s) 1, (?m) 1, dynamic 25.
There are no fixed \p{...}, \z or (?P<name>) sites in this source pin. Their accepted
Go-option spellings are nevertheless exercised by the dialect gap controls.
82 fixed patterns produce 125,771 whole-match records, 2,354,870 bytes identical
on all four paths over 2,104 inputs from the documented alternate 100-file corpus.
The Go oracle runs under cohere's own pinned Go module and maps byte spans to UTF-16
at one observation boundary. No per-rule byte conversion or matching fallback exists.

The native constructor gap and five observed raw-option dialect disagreements are
recorded for @system_adamic_library in gaps.md. None of the three requested rule
migrations is claimed complete. Existing no-warning-comments source was preserved;
the shared harness and registry were not edited. Two fixed comment patterns are
exported from patterns.a for integration after the shared harness is ready.

Every fixed site's translation agrees on this bounded corpus. No fixed site was
observed to require an unavailable JS construct. An unrestricted raw Go option
pattern has no faithful translation under the required unchanged new RegExp(pattern,
'u') contract: \s disagrees, and (?i), \p{Greek}, \z and (?P<name>) are rejected by
Node. This does not mean those expressions lack port-time JS translations; it means
port-time translation cannot operate on an unknown future configuration string.

The pinned cohere 715ba94f has 39 regexp-importing files, 88 MustCompile calls and
19 Compile calls. Current upstream 7945d102 was inspected separately without
changing the submodule: it had 89 MustCompile calls. Neither census has 92.
Six pinned MustCompile sites are constructed sources; all 25 dynamic compile sites
are preserved with their Go source expression and the required runtime contract.
These source expressions include option families and internal generated templates;
an infinite accepted pattern language cannot be enumerated as finite literal rows.

Toolchain setup completed: Go 0s, clang 0s, Node 0s, submodules 16s,
build cache warm/done 388s; nproc 5, cgroup CPU quota four, reported memory 17.6 GB.
All test output is logged under evidence/. The first emitted-JS test lacked the
runtime resolver; the test now executes emitted output through oracle/node.mjs.
A preliminary dialect assumption about a trailing LF was corrected to Node's
observed false result; the final gap test records the actual observations.

Exact validation:

```
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/regex -count=1 -v -timeout=20m > /tmp/regex-final.log 2>&1
go vet ./stage1/cohere/lint/regex > /tmp/regex-final-vet.log 2>&1
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^regexp\.a$' -count=1 -v -timeout=10m > /tmp/regex-filtered-oracle.log 2>&1
```

## Seat correction

Merged area d45a323be and rebuilt at cohere 7945d102. Census remains 107:
new locale-stem pattern in bb39e2dd, removed Tailwind Go regexp site in 0f2797dd
(replaced by esregexp), zero pattern edits, 66 line-only moves. Complete provenance
is in testdata/shapes/census-review.json and CENSUS.md.

Normal go test now executes every fixture through shapes_test.go. Full regex package
passed in 64.454s, logging 107 Node/JavaScript comparisons and 107 named native
refusals. The separate mutant-enabled go test failed only at its fixture comparison.
Vet is clean. No unexplained deletion and no runtime/compiler or migration edit.
