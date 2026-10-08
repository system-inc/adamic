# c2e39b75: no-warning-comments

For @system_cohere_lint parity review. Based on `origin/cohere-pin/f5d1934a-wip`
(`84e92ceac3e8597f63749bc1b7718dade9629549`), cohere pin
`f5d1934a2d7bebe706210cb1cfd01aebff4f8ca7`. Upstream change:
`c2e39b755570a15c8c2bd39aec64314142003a34` (#7mztrdd).

The affected port is `../../comments.ts`, consumed by this rule alone. Decorations
are literal characters, including hyphens; the start prefix, quoting and directive
trimming use JavaScript's 25 whitespace characters (FEFF included, NEL excluded).
JavaScript `iu` word boundaries include the ASCII folds of long s and Kelvin sign.
The matcher uses supported scalar comparisons; no language gap was encountered.
Cohere source and every existing check are preserved. The decoration mutant now
uses the upstream `+` witness and restores range interpretation.

## Upstream fixtures added or changed for this rule

All results compare complete Go output with source Node, emitted JavaScript and
ASan/UBSan native, including messages, byte ranges and fixed sources.

| Added rule fixture | Options | Findings | Result |
| --- | --- | ---: | --- |
| `//\u00a0TODO later` | default | 1 | PASS |
| `//\ufeffTODO later` | default | 1 | PASS |
| `/*- todo */` | decoration `['*','-','/']` | 1 | PASS |
| `/*+ todo */` | decoration `['*','-','/']` | 0 | PASS |

The six added lead values are, in column order, `\u00a0todo`, `\ufeff todo`,
`\u2028todo`, `\u3000 fixme`, `- todo`, `-+ todo`. Every cell below passed;
values are finding counts. Each configuration is upstream's exact configuration.

| Lead configuration | NBSP | FEFF | LS | ideographic space | hyphen | hyphen-plus |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| default | 1 | 1 | 1 | 1 | 0 | 0 |
| decoration `['*','/']` | 1 | 1 | 1 | 1 | 0 | 0 |
| changed decoration `['*','-','/']` | 1 | 1 | 1 | 1 | 1 | 0 |
| decoration `['k']` | 1 | 1 | 1 | 1 | 0 | 0 |
| terms `['*todo',' fixme','kelvin','ſtop','日本']`, decoration `['*']` | 0 | 0 | 0 | 1 | 0 | 0 |
| terms `['']` | 1 | 1 | 1 | 1 | 1 | 1 |
| anywhere | 1 | 1 | 1 | 1 | 1 | 1 |

The changed hyphen configuration also passes on all 31 pre-existing UTF-8 lead
values. The complete [fixture list with results](evidence/upstream-fixtures.tsv)
names each value and its configuration; [focused output](evidence/focused.log.gz)
records all 161 row results and mutant witnesses. The upstream direct
lead test, including its unchanged invalid-byte value, passes in Go. All upstream
`TestNoWarningComments*` tests pass ([upstream.log](evidence/upstream.log.gz)). The
port's ordinary upstream capture holds every asserted rule case to all backends.
The additional checks cover all 25 whitespace characters and NEL in prefix,
quoting and directive consumers, plus six folded-boundary cases: 161 rows total.

## Case 108

`generated-10.ts`, decoration `['z','-','a']`: before **0 findings** (descending
RE2 range dropped the matcher); after **4 findings**, byte-identical to new Go:

| Term | Line:column | Byte range |
| --- | --- | --- |
| todo | 1:1 | 0–21 |
| fixme | 2:79 | 100–111 |
| todo | 3:1 | 170–189 |
| fixme | 4:37 | 226–249 |

Fixed source is unchanged. Exact outputs: [before](evidence/case108-before.txt),
[after](evidence/case108-after.txt). Before uses an isolated copy with the base
matcher; after is additionally compared directly with the pinned Go oracle.

## Validation

Complete lint package: **116 passed, 0 failed, 1 skipped**, counting named tests
and subtests; package exit 0, 1905.148 seconds. The unchanged skip is
`TestCheckerBridgeRefusalPending`, awaiting cohere's checker error-as-value API.
Both former reds, `TestRulesAgree` and `TestProfileSnapshotsAgree`, pass.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
ADAMIC_LINT_BENCH=1 \
ADAMIC_LINT_PROFILE_DIR=/workspace/scratch/lint-c2e39b75-profile \
ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/scratch/lint-c2e39b75-profile \
go test ./stage1/cohere/lint -count=1 -json -timeout 90m
```

Evidence: [complete package log](evidence/package.jsonl.gz). Upstream capture:
3,883 unique combinations; aggregate comparison 13,777,247 identical bytes.
Compiler/stage-one corpus: 872 files, 30,407,140 identical bytes. Profile snapshots:
43,482,883 identical bytes. Five throughput rounds each report 28,312 findings on
Go, native and Node. All registered mutants, owned witnesses and shard checks pass.
The three warning-comment mutation controls (range interpretation, Go-five
whitespace, complete base reversion) each compile, run, and differ from Go on
source Node, emitted JavaScript and sanitized native.

`go vet ./stage1/cohere/lint/...`, `gofmt -l` on both touched Go test files, and
`git diff --check` pass with no output. Cohere's working tree is clean.
