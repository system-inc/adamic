Built: removed the hand-written commentPattern matcher shared by default-case and no-fallthrough; fixed default-case regex uses the shared-table RegExp literal.
Commits: prior rule delivery 551c529c8 and helper delivery 65890c010 are based on current main c01907a7; migration SHA is in the response.
Checks: 32 fixed-pattern queries, private Go/source Node/emitted JavaScript/sanitized native identical (184 bytes); required dynamic constructor refuses lowering; latest setup 83s, nproc 5.
Mutants: absolute_end_removed compiles and runs successfully on all three runtimes, caught only by private-Go byte comparison; option_parity_guard_removed caught on source Node only.
Limits: default-case/no-fallthrough runtime options are blocked; no current full-rule native/emitted parity, no new helper claim, no full gate.

Current main is b8fb957aa839a9e8cb0b54279dd9864fa317bd30. ab70f38d47de1d4974082b38f84a56af2368b7af (origin/lint-rules/harness) is not its ancestor yet. Both owned branches already include main; no rebase onto an area branch or unrelated harness branch was performed. When that harness lands on main, rebase and re-green the rule branch before further claims. The previous parking exception is insufficient now because the regex compiler/dialect blockers are additional obstacles.

The fixed literal uses codex/lint-regex table row core/default_case.go:29 unchanged: /^no default(?![\s\S])/gui. The shared row transforms Go's end anchor into an absolute-end assertion. Each independent test resets lastIndex. Tests call the actual private Go defaultCaseDefaultCommentPattern through an export-only overlay. The mutant removes the absolute-end assertion, accepts suffixes, compiles and finishes with exit zero/empty stderr on all three runtimes, then fails only output comparison.

Runtime options now use new RegExp(source, 'u'). The old literal/anchor/character-class switch and custom case-folding matcher are gone. Empty options remain the existing rule-default path. Pattern.match explicitly refuses nonfixed options until Go parity is resolved, so supported JS syntax cannot silently replace different Go matching behavior. Invalid JS constructor syntax throws rather than pretending to reproduce Go's invalid-pattern fallback. Those configurations are blocked, not successful findings parity.

The independent native/emitted reproducer is testdata/runtime-pattern.a. With argument TODO, source Node prints true and exits zero. Both adamic build and adamic js refuse with: stage 0 can't lower RegExp with a nonconstant pattern yet. The same constructor prevents lowering these complete rules even for inputs that happen to select the default. Changing internal/lower/regexp.go or adding a matcher fallback is outside this unit and contrary to the new rule.

Even after that compiler gap is closed, raw options need a dialect decision. Concrete external Go/Node observations:

| Go source | Required JS source | Input | Go | Node |
| --- | --- | --- | --- | --- |
| \s | \s | U+00A0 | false | true |
| (?i)todo (no-fallthrough prefix) | todo | TODO | true | false |
| (?i)todo (raw option) | (?i)todo | TODO | true | SyntaxError |

No runtime translation or hand-written matching is substituted. These differences cannot simultaneously satisfy arbitrary Go options and the specified raw JS constructor. The Node-only refusal guard mutant accepts a|b after removing the guard; its successful output differs from the named baseline refusal. Native/emitted cannot compile that mutant, so those runtimes are not credited with killing it. The previous unsupported_pattern_accepted gate and full frozen rule results are historical and superseded for the affected two rules.

Reproduce from the rule branch with a current-main compiler checkout at /workspace/adamic:

```sh
source /workspace/adamic-tools/env.sh
# In the current-main compiler checkout:
go build -o /tmp/wave10-regex-adamic ./cmd/adamic
# In the rule checkout:
python3 stage1/cohere/lint/rules/default-case/regex_contract_validate.py > /tmp/wave10-regex-contract.log 2>&1
```

Every command's stdout/stderr is retained under evidence/regex-contract-*. Native successful comparisons use ASan/UBSan. The validator succeeds when the supported fixed comparison/mutant and the documented blockers are observed; it is not a green full-rule oracle. No finding.ts, context.ts, main.ts, generator, shared oracle or lint_test comparison was edited. The announced leak-helper diff is absent from main on this fetch; nothing was reverted. All new Adamic files are .a.

Earlier landing-cap refresh: origin/main was still c01907a7 and does not contain ab70f38d4. Both owned branches already have that main as an ancestor. The regex contract validator was rerun: fixed/private-Go 32-query parity and the three-runtime compiling absolute_end_removed mutant pass; the required runtime constructor still refuses both emitted and native lowering. External Go/Node dialect witnesses and the Node-only option guard mutant reproduce unchanged. The full rule oracle remains blocked, so this branch is not landing-ready and no new helper is claimed. See evidence/regex-refresh-summary.log.txt and evidence/regex-refresh-setup.log.txt.

Listener contract correction: the twelve owned rule.json descriptors already declare ast.Kind names as strings. They remain the registration source of truth. The earlier unused numeric listeners.a probes are historical evidence, not a dispatch contract, and no numeric declaration is proposed for the unified harness. Existing index visitors continue through the registry's compatibility path; no shared driver or comparison file was changed. No leak-helper replacement has reached this main snapshot; none was reverted.

Current landing-cap refresh: rebased cleanly onto main b8fb957aa839a9e8cb0b54279dd9864fa317bd30, including its inherited static-field emitter fix. Rebuilt the compiler from that base and reran regex_contract_validate.py. The 32-query fixed literal matches private Go on Node, emitted JavaScript and sanitized native; absolute_end_removed compiles and is caught only by comparison on all three. The runtime constructor still refuses both lowering paths, the raw option dialect differences reproduce, and option_parity_guard_removed is credited only on source Node. This validates the documented blockers, not a green full-rule gate. No new helper or rule claim is made. The separately rebased helper branch passed its complete four-helper oracle in 55.364s with all four compiling mutants and vet. Setup total 83s, nproc 5. See evidence/regex-b8fb957a-summary.log.txt. Shared harness ab70f38d4 remains outside main; no area/main push or shared-file edit was performed.

After the requested landed-harness rebase, current area is 7481e0324 and current main is 39638d9e2. The unified TestRulesAgree now loads the .a rules and fails in 73.188s at default-case/pattern.a:13:96 on the required nonconstant RegExp constructor. The compiler/dialect blockers therefore remain after shared harness integration. See DEDUP.md and evidence/area-7481e032-unified-oracle.log.txt. Fixed-literal three-runtime parity and its compiling mutant were revalidated with a newly built area-based compiler; no current full-rule green is claimed.
