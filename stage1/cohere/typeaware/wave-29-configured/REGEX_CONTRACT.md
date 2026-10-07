Built: rebased wave 29 onto main c01907a7, removed its handwritten regex VM and routed id-match option matching through new RegExp(pattern, 'u').
Commits: rebased source e9d4af03; final implementation/evidence sha in handoff; only codex/typeaware-wave-29 is pushed with its exact prior-tip lease.
Commands: setup 37s, nproc 5; default original and completed continuation rules, JSX kernels, graph kernels and metadata re-green; required runtime regex source is explicitly refused natively.
Mutants: default rule/provenance/released-handle, three continuation rules, three JSX kernels, four graph kernels, eighteen metadata and a constant-option substitution caught by independent comparisons/refusals.
Uncovered: configured id-match parity is blocked by dynamic RegExp lowering and Go/JS option dialect; JSX source integration and context capture/memo analysis remain incomplete; no new claims.

The updated user instruction bans handwritten regex matchers. The former native
RegexpProgram VM is therefore deleted, not kept as a fallback. id_match.a no
longer imports it. The configured rule-local profile calls optionPattern(pattern)
and passes expression.test(name) to the existing native judgment/report logic.
option_pattern.a implements exactly new RegExp(pattern, 'u'), matching the shared
codex/lint-regex option constructor contract. Its inventory row and branch sha
are recorded in regex-contract-validation. This is an arbitrary option pattern,
not a fixed pattern that can use a port-time literal translation row.

Default id-match remains empty-pattern behavior. Invoking its legacy run method
with a nonempty pattern now refuses explicitly; the configured profile uses the
separate required constructor path. No handwritten matcher remains in these
native rule/regex runner sources. The old bridge raw regexp-program question is
retained as an unused historical ABI fact question, not used by active lint.
The earlier REGEXP_REPORT.md and VM comparison scripts are historical proof of
the superseded implementation; they do not establish current configured parity.

## Proven blockers

Source Node executes the runtime option probe and prints true for TODO. Current
native compilation refuses both that probe and the actual configured profile:
`stage 0 can't lower RegExp with a nonconstant pattern yet`. The probe's pattern
comes from programArguments; a literal constructor control does not prove dynamic
support. The constructor logic is isolated so the default profile remains buildable.

A mutant replacing the runtime pattern with the constant TODO compiles and exits
zero with empty stderr. Input NO yields false on the unmodified Node option path
and true on the mutant native artifact. Only the independent runtime-option
comparison catches it; it is not killed by a compiler error. This proves the new
option-source check, not complete configured findings.

The actual owned option factory was also run on Node against the preserved Go
regexp oracle. Input strings and pattern strings are serialized as JSON to keep
backslashes/newlines exact. Observations:

| Pattern/input | Go | JS constructor with u |
| --- | --- | --- |
| \Afoo\z / foo | true | SyntaxError |
| ^foo$ / foo followed by LF | false | false |
| . / CR | true | false |
| (?i)^foo$ / FOO | true | SyntaxError |
| ^[^_]+$ / foo | true | true |

The first observation run used incorrect JS string escaping and was corrected
before final evidence: the final probe asserts all five exact outputs. There is
no runtime translation or matcher fallback. Configured id-match cannot currently
satisfy both the mandated raw JS option contract and arbitrary Go dialect parity.
The shared regex branch reports the same fleet-level dialect issue. A changed
option contract/restriction is not assumed here. Configured matching is blocked,
not silently miscompiled or claimed complete.

## Landing and verification

Main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. All 21 own commits
rebased cleanly. No developer-tools leak-helper changes were reverted or fought;
none required conflict resolution in this rebase. A fresh compiler was built as
/workspace/wave29-regex-contract-adamic, and owned checks selected it explicitly.
Main stayed unchanged at the final remote check. No main/area push or PR.

- Original TestWave29AgreementAndMutants PASS 102.536s: 41 controls, 15 findings,
  10250 bytes (path headers changed); controls and frozen compiler77/5241 bytes
  and repository287/18485 bytes match Go under sanitizers. Default denylist,
  match, race-range and provenance mutants differ with exit zero. Retained-
  released-handle mutant is caught by required panic. Empty-pattern id-match
  parity is covered, not configured matching.
- Completed continuation: 400 configured cases, 252 findings in seven profiles;
  both frozen corpora match Go, including full native/bridge sanitizers. Globals,
  setter and shadow mutants compile/exit zero/empty stderr and differ.
- JSX fourth batch: 79 rows/7359 bytes, Go/native/Node/emitted JS/sanitized native
  identical; import-module, intrinsic-dash and conditional-order mutants caught.
  All three Go source witnesses still report while the native JSX parser
  misclassifies fragments or refuses provider/undefined self-closing syntax.
- Static kernel: 34 supplied graphs/27 findings/14419 bytes, two compiling
  mutations caught; sanitizer agreement. Control kernel: 17796 graphs/243119
  bytes, Node/JS/native/sanitizer agreement, throw/switch mutants caught.
- Nine listener manifests/declarations: 315 Go/native/Node/sanitizer bytes agree;
  nine JSON and nine compiling native numeric mutations caught.

Whole-process timing observations while validation jobs ran concurrently:
original compiler native1.896475576s/Go0.306935433s (6.18x), repository
native0.273928873s/Go0.124837782s (2.19x). Continuation compiler native2.213451748s/
Go0.416732655s (5.31x), repository native0.950845702s/Go0.380764718s (2.50x).
These are existing completed/default profiles, not configured regex or full JSX
lint performance. No new claims are taken while current claims remain blocked.

No full repository gate, full configured naming corpus, legacy VM/raw-instruction
suite, standalone new lifetime suite or full JSX findings/fixes/suggestions parity
was run. The removed VM's historical broader regex guarantee no longer applies.
Graph and JSX helper comparisons remain partial-source coverage.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-regex-contract-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v > /tmp/wave29-regex-contract-default.log 2>&1
ADAMIC_COMPILER=/workspace/wave29-regex-contract-adamic python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-regex-contract-next > /tmp/wave29-regex-contract-next.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/check_regex_contract.py /workspace/wave29-regex-contract-proof-final > /tmp/wave29-regex-contract-proof-final.log 2>&1
```

The other owned script command records and full logs are preserved alongside
regex-contract-validation; all output went to logs. Setup tools/submodules 0s,
cache warm/total37s; nproc5, CPU quota four cores, 17.6 GB. Native code is .a.
