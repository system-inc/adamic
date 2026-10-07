Built: verified the named shared harness and narrowed the remaining JSX source-port blockers; no production rule changes.
Commit: implementation parent 000793ab89f08f829f60e5dbb65459da7ce6c230 remains rebased and green on current main c01907a7; this report is pushed only to codex/typeaware-wave-06.
Checks: parser from ab70f38d47de1d4974082b38f84a56af2368b7af parses all three claimed JSX shapes; native and ASan/UBSan match source Node on counts 19, 9, 18 with empty stderr.
Mutant: omitting parser.file() compiles and exits zero; comparison with original source Node bytes catches it.
Not covered: shared numeric node kinds, raw checker/declaration adapter and native React source analysis; no new claims or full source parity.

A fresh all-heads fetch finds main unchanged at c01907a7 and the own remote unchanged at 000793ab. The named harness is not an ancestor of origin/main yet; it is being integrated into area/stage1-lint. It was examined in a detached scratch worktree, without modifying its files, importing its history into the worker branch or pushing any main/area ref. The existing owned production inputs and their post-rebase oracle evidence remain unchanged.

The harness adds reportNode/reportRange and suggestions beside the original seven-argument report method. Its tests demonstrate .a module support. JSX syntax parsing is now available on this branch. The previous broad statement that JSX parsing is missing is narrowed: it remains absent from current main, but is present and proven usable on this named branch. A standalone probe imports that parser and parses React.Fragment, an undeclared component tag and adjacent intrinsic children. It does not perform any lint analysis. Normal native, sanitized native and source Node agree, with empty stderr. The parse-omission mutant exits zero and changes only the observable node counts.

The remaining source-port blockers are explicit. ParseNode still exposes kind: string, not a numeric SyntaxKind. The harness's RuleContext is syntax-only and declares no checker, binder or symbol lookup. Its example JSX rule still accepts an index and fetches the node. Its example kinds array contains strings. The owned rules must honor the stricter numeric-kind and handed-node instruction; they cannot adopt string comparisons or per-rule node refetch to force integration. Private raw declaration/initializer/symbol-origin fields still need a production adapter. The reporting model itself is no longer a blocker, and the tested prepared-node kernels do not need a finding-model rewrite.

The three JSX claims remain reserved, tested and pushed as partial prepared-node ports; their source entry points explicitly refuse until these remaining contracts land. Earlier React HIR claims stay parked on #dnv6f2c's source analysis dependencies. No new claims are taken, and no shared registration, harness, parser, compiler or bridge file is edited. Developer-tool leak changes are not present in this unchanged main snapshot and nothing is reverted.

Reproduction uses the retained stage-0 artifact whose compiler inputs are unchanged, with a detached worktree at the named harness SHA:

```sh
git worktree add --detach /workspace/wave-06-harness-review ab70f38d47de1d4974082b38f84a56af2368b7af
/workspace/wave-06-landing4-jsx/adamic build /workspace/wave-06-harness-probe.a -o /workspace/wave-06-harness-probe
/workspace/wave-06-landing4-jsx/adamic build /workspace/wave-06-harness-probe.a -o /workspace/wave-06-harness-probe-asan --sanitize
node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/wave-06-harness-probe.a
```

Actual build/probe output is redirected directly into retained files. Probe and mutant source, outputs and hashes are in harness_evidence. These are parser availability observations, not source rule finding comparisons or end-to-end timing measurements. Prior complete rule, bridge ownership and prepared-node comparison evidence remains in LANDING4_REPORT.md.
