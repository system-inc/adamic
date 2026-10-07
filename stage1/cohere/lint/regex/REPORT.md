## Review correction and parking

The fleet ruled that user option patterns are JavaScript, not Go regexp. The old
raw-option dialect comparison below is historical and is superseded by this section.
The option test now runs cohere esregexp Compile(source, "u") against Node's unchanged
new RegExp(source, "u"), using nine agreement controls and one separately named
property-gap control. The constant-pattern table, literal module and comparison
corpus are unchanged from 071fb0128.

Both parking dependencies are named here and in gaps.md: the dynamic RegExp library
reproducer, already sent to @system_adamic_library by the fleet, and the Go option-rule
migration #7mztrdd. The pinned esregexp property escape `\p{Script=Greek}` is also
observed to reject a pattern Node accepts; its reproducer remains explicit and is
excluded from agreement totals. The three rule migrations remain waiting, with no
fallback. This branch is parked until these prerequisites clear.

Review validation is recorded in evidence/options-review.log and options-vet.log.
Only the changed option/gap tests were rerun; the unchanged constant-pattern gate
and its three-backend mutant retain the earlier evidence below.

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

The native constructor gap is recorded for @system_adamic_library in gaps.md.
The former RE2 option comparison is superseded by the review correction above. None of the three requested rule
migrations is claimed complete. Existing no-warning-comments source was preserved;
the shared harness and registry were not edited. Two fixed comment patterns are
exported from patterns.a for integration after the shared harness is ready.

Every fixed site's translation agrees on this bounded corpus. No fixed site was
observed to require an unavailable JS construct. The user-option contract is JavaScript and will be enforced by cohere esregexp after
#7mztrdd; Go-only input syntax is not the option contract.

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
