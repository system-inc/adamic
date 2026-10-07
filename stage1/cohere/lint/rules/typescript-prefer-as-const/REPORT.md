Built: three claimed TypeScript repair-rule candidates, retaining every fix and suggestion in .a.
Commits: claim 058844c1; constraint c48a45cd; as-const 7ce1f1f7; enum fd018b4c; evidence in git history.
Commands and outputs: final corpus, backend, mutant and extra-witness suite; registration and vet logs in evidence.
Mutants: unknown constraint suppressed; insertion changed to as mutable; only the second enum suggestion changed. All are comparison-only witnesses.
Not covered: standard frontend integration is blocked; no shared files edited, no further claims, no full gate pass.

The three rules are @typescript-eslint/no-unnecessary-type-constraint, @typescript-eslint/prefer-as-const and @typescript-eslint/prefer-enum-initializers. The helper-ready list was exhausted at selection. Claim 058844c1 was pushed before code, after fetching every origin branch and checking main and 39 distinct claim Markdown blobs. Main was ef3d907ecdc4c771b016f7d9c52372def057a340. The syntax-ready inventory supplied these first three remaining rules.

The rule-owned diagnostic model preserves diagnostic ranges, all safe fixes, all suggestion IDs/messages, and all edits in each suggestion. The independent oracle calls unmodified Go cohere rules; only its serialization exposes all records instead of discarding repairs. The Adamic driver prints exactly those records and applies all safe fixes to the source. Original Go test assertions also execute outside Adamic. Filenames retain .tsx/.mts/.cts where those alter generic-arrow suggestions. UTF-16 offsets are translated into Go's UTF-8 byte offsets.

These candidates expose their complete records via analyze. They remain registered through rule.json, but legacy visit explicitly refuses incompatible repair shapes. The constraint suggestion is outside the diagnostic range, as-const has multiple fixes, and enums have three suggestions. Silently reducing these records would violate the byte-for-byte requirement. This is an integration blocker, not a claim that the standard frontend works. The whole unit must be merged together because the diagnostic model and verification driver live in the as-const directory.

Unmodified shared harness command: `go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1`, exit 1: `profile_test.go:32:23: cannot range over portFiles` (the value is now a function). A direct registered-visitor probe exits 70 with `no-unnecessary-type-constraint requires complete repair renderer`. Shared registry currently assumes rule.ts and shared copying assumes .ts. Another worker owns codex/lint-harness-dot-a. No shared source was edited. Per Ahra's correction, stop here after pushing the owned candidates and evidence.

For bounded verification, scratch Go overlays load .a, correct the preexisting profiling compile error and add a rule-owned comparison driver. All overlay sources are saved as text under this rule's evidence directory; they are not production edits and are not offered as an integration patch.

Commands (each writes directly to a log):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -overlay=/tmp/lint-wave1-13-repairs/overlay.json ./stage1/cohere/lint -run '^TestWave13Repairs' -count=1 -v -timeout 20m > /tmp/lint-wave1-13-repairs/parity-final.log 2>&1
go test -overlay=/tmp/lint-wave1-13-repairs/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-13-repairs/registry-final.log 2>&1
go vet -overlay=/tmp/lint-wave1-13-repairs/overlay.json ./... > /tmp/lint-wave1-13-repairs/vet-final.log 2>&1
```

Reproduce from a fresh scratch directory: `python3 evidence/reproduce.py --compiler /path/to/typescript-6.0.3`. The pinned compiler checkout was 050880ce59e30b356b686bd3144efe24f875ebc8.

Toolchain setup used the existing scratch compatibility overlay to get past the profiling error: `GOFLAGS=-overlay=/tmp/lint-wave1-13/overlay.json bash cloud/setup.sh`. Timing: Go 0s, clang 1s, Node 1s, submodules 1s, build cache 20s, total 20s. nproc=5; cgroup quota 400000/100000; 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

An extra mapped-type witness initially failed comparison: Go reports `[K in any]` using the type-parameter listener, despite the extends-focused message. The correction recognizes InKeyword as well as ExtendsKeyword. The failed extra-1.log is retained as evidence of the independently detected error. It is not a passing test.

Final measured results follow below.

Corpus: 77 compiler files and 148 stage1 .ts/.a files, plus 43/69/21 upstream vectors and one owned witness per rule. This gives 269/295/247 rows, 811 comparisons total. All selected inputs are supported: zero parser gaps for these three rules. The capture tool additionally reports an inherited return-void JSX gap outside this unit; that case is not counted as covered here.

Throughput is best of three interleaved whole-process count runs, including startup, parsing and I/O. Native uses a separate unsanitized release binary for timing; parity and mutant execution use sanitized native. Counts are identical across Go, native and Node. These are measurements of this manifest and driver, not an isolated algorithm benchmark.

| Rule | Rows | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: | ---: |
| no-unnecessary-type-constraint | 269 | 28 | 21.63 | 26.43 | 122.97 |
| prefer-as-const | 295 | 27 | 22.07 | 26.35 | 123.94 |
| prefer-enum-initializers | 247 | 702 | 591.93 | 692.79 | 3178.13 |

Mutation details: removing UnknownKeyword recognition drops the unknown-constraint diagnostic; replacing the inserted ` as const` with ` as mutable` changes the second safe fix and resulting source; changing `${position + 1}` to `${position + 2}` changes only the second enum suggestion while leaving the first suggestion intact. Each mutant must compile, run successfully with empty stderr and exit zero on all three backends before comparison rejection is credited. Thus the enum mutant also proves the comparison examines more than the first suggestion.

The corpus suite passed in 192.02s and the mutant suite passed in 40.02s. Serialized corpus totals: 12,388,093 / 12,386,711 / 12,774,066 bytes, 37,548,870 bytes overall. The extra-witness suite separately verifies the mapped-type correction against Go. Registration passed in 0.032s and vet produced an empty successful log. Raw logs include the intentional default-harness failure and legacy renderer refusal; these are not counted as passes.
