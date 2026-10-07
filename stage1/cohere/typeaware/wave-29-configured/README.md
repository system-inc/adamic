The regexp blocker below is resolved by [the native instruction VM](REGEXP_REPORT.md). This file preserves the earlier observation and validation scope.

Built: configured id-denylist and the id-match AST judgments, flags, spans and messages.
Commits: baseline 0ab90614; configured implementation f114b002.
Commands: rule-local check.py PASS on 246 inputs; existing wave gate PASS 72.787s; vet and gofmt exit 0.
Mutants: both configured rule mutations differ only through output bytes; existing default-rule, provenance and released-handle mutants also caught.
Blocked: general native Go-compatible regexp execution and emitted-JavaScript checker calls; no additional rules claimed.

This follow-up replaces the earlier report's missing configured denylist checks.
It does not claim that id-match is fully ported. No shared harness, registration
generator, compiler, parser or bridge file changed in this follow-up. New Adamic
files, including input and mutant copies, use `.a`. The branch remains
`codex/typeaware-wave-29`.

## Implemented and measured

`../id_denylist.a` accepts the configured name list, follows the production
callee/import/destructuring/property-write/import-attribute exclusions, and
exempts a resolved symbol only when all declarations are declaration-file
nodes. A nil symbol is checked rather than mistaken for a global. The existing
symbol-provenance question supplies raw records; native code decides every
finding. Private names match without `#` and render it in the message. Core
rule names have no `@typescript-eslint/` prefix.

`../id_match.a` implements the production AST decision order, local imports,
class-field and private-method distinction, property reads versus writes,
declaration gates, and the distinct binding and assignment destructuring
branches. `runConfigured(pattern, matches, flags)` explicitly requires a matcher
callback. All four flags and native findings are implemented. `run()` still
handles the inactive production defaults; giving it a nonempty pattern panics
with its missing-matcher explanation. No pattern is silently treated as a
successful match.

The isolated driver uses a test-only exact implementation of `^[^_]+$`, checking
nonempty names without underscores. This is not a general regexp engine and
not an implementation of arbitrary configured patterns. Its inputs come from
unchanged production Go test tables via Go's AST, including additional test
tables beyond the original fixture collection. Two Unicode/CRLF controls are
added. Sources are unchanged apart from a final `export {};` matching the
production fixtures' module scope and preventing unrelated files from merging
local declarations in the single checker program. The Go oracle independently
loads that same program and calls the unmodified production rules with their
actual option decoders. It does not import the bridge or native decisions.

| Population | Inputs | Findings |
| --- | ---: | ---: |
| Configured denylist | 148 | 132 |
| Configured match with `^[^_]+$` | 98 | 62 |
| Total | 246 | 194 |

All three configured streams, Go, native and ASan/UBSan/LeakSanitizer native,
are exactly 65,481 bytes, SHA256
`a83bc63a4e05ee2f94b1cc3a50eb0f38578330b0ea6f6413348e1356b55cfb73`.
Native stderr is empty. The canonical serialization includes every finding,
span, message, fix and suggestion field; fixes and suggestions are all zero as
in production. Compressed streams, raw fixture inputs, command exits/times and
hashes are under [validation](validation/).

The configured denylist mutant returns false at its final should-check gate.
The configured match mutant removes its AST eligibility gate. Both compile,
exit 0 with empty stderr, and produce different diagnostic bytes. Their exact
first differences and complete mutated streams are preserved. Neither is
counted as caught by compilation or sanitizers.

The unchanged existing wave gate was rerun against the final implementation:
41 default controls, 15 findings and 9,922 equal bytes; the frozen 77 compiler
roots, zero findings and 5,241 equal bytes; and the frozen 287 repository roots,
zero findings and 18,485 equal bytes. Every population also agrees under native
sanitizers. The three default-rule mutants, provenance-output mutant and
released-handle registry mutant are caught. Querying a released handle retains
its required panic. TypeScript remains v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`; manifests and pins were unchanged.

The first regression attempt correctly refused the match mutant because the new
configured method repeated the harness's exact default-return marker. The
configured method now tests `pattern.length === 0`, preserving that marker
uniquely without a harness edit. That failed log and the final passing log are
both saved. An initial regexp probe failed for an unrelated boolean console
argument; it was corrected to print strings before measuring the actual gap.

## Observations and remaining blockers

`gaps/regexp.a` asks for `new RegExp('^[^_]+$', 'u')` and prints its test result
as a string. Native compilation exits 1 with:

```text
adamic: .../gaps/regexp.a:1:17: stage 0 can't lower new an Identifier yet
```

The current branch has no native regexp lowering/runtime. Implementing the
Go RE2 pattern surface or integrating the separate regexp compiler work would
require work outside these rule files. A JavaScript RegExp engine alone would
also need to be checked against Go's accepted syntax and semantics. The supplied
callback is an explicit integration point, not evidence of that missing engine.
General configured id-match execution therefore remains unfinished. Per the
instruction to stop on other blockers instead of editing shared files, no next
three rules were claimed.

Emitted JavaScript for the checker-dependent local driver also exits 1:

```text
adamic: .../runner.a:18:17: Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>
```

This command has no JavaScript checker adapter on this branch. The CLI also
contains an explicit JavaScript/external-checker refusal. After fetching all
origin heads, `origin/codex/lint-harness-dot-a` was not present in this checkout.
No emitted-JavaScript agreement is claimed for these new checks. Shared profile
compilation and suggestion serialization were not edited.

One quiet whole-process observation on the configured controls was native
0.071319 s versus Go 0.094750 s, 0.75x. This is one observation rather than a
benchmark median. Default compiler timing was native 1.679721 s versus Go
0.273064 s, 6.15x; repository timing was native 0.252398 s versus Go 0.114338 s,
2.21x. Native remains slower on the frozen default corpora. Timing includes
complete output. No other build or test from this worker ran during those
observations.

## Reproduce

Setup passed in 17s; go, clang, Node and submodules each reported 0s, cache warm
17s. Go 1.27.1, clang 20.1.8, Node v24.19.0; `nproc` 5, cgroup cpu.max
`400000 100000`. Setup output is saved in validation/setup.log.

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-configured/check.py \
  /workspace/wave29-configured-final > /tmp/wave29-configured-final.log 2>&1
# Exit 0: 246 controls, configured mutants, sanitizers and RegExp gap refusal.

ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-followup-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave29-followup-final.log 2>&1
# PASS, 72.787s.

/workspace/wave29-configured-final/adamic js \
  stage1/cohere/typeaware/wave-29-configured/runner.a \
  > /workspace/wave29-configured-final/javascript-gap.stdout \
  2> /workspace/wave29-configured-final/javascript-gap.stderr
# Expected gap: exit 1, unlinked checker library refusal.

go vet ./stage1/cohere/typeaware ./bridge/tsgo/... \
  > /workspace/wave29-configured-final/vet.log 2>&1
# Exit 0, empty output.
gofmt -l stage1/cohere/typeaware/wave-29-configured/testdata \
  > /workspace/wave29-configured-final/gofmt.log 2>&1
# Exit 0, empty output.
```

The full repository gate, arbitrary configured regex patterns, every upstream
case/options matrix, JSX/JavaScript corpus populations, and emitted-JavaScript
comparison remain unverified. No pull request was opened.
