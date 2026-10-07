# Wave 1 slot 07 completion and next queue audit

Built the three original claimed .a rules and the remaining Google Font Display selector/listener.
Pushed cb2dde25, f66a15c3, 6685e5bd, f59d597f, followed by adapter formatting f2ef276d.
Owned validation passes Go/Node/emitted JavaScript/sanitized native; setup recovered and completed in 20s, nproc 5.
Four compiling semantic mutants are rejected only by Go comparison; Google Font validation is selector-only.
Full Google Font JSX parity and whole repository gate are not covered; no eligible ready rule remains after origin-port skips.

The original three ports are complete under the documented comparison contract. Each has byte-identical findings, messages, ranges, suggestions/edits where present, and unchanged automatic-fix output over upstream source literals and TypeScript src/compiler plus stage1 sources. Final corpus is 351 file rows, including all modules added in this completion. Exact commands, raw output, mutants and rates are in each owned rule's README.md and evidence/ directory.

| Rule | Upstream unique source rows | Final corpus identical bytes | Native findings/s | Node findings/s | Go findings/s |
|---|---:|---:|---:|---:|---:|
| no-unsafe-negation | 62 with both ordering settings | 13,278,673 | 92,265.76 | 8,990.31 | 198,183.80 |
| no-unsafe-optional-chaining | 103 with both arithmetic settings | 13,278,673 | 74,697.71 | 9,027.52 | 215,207.05 |
| no-underscore-dangle | 93 with valid upstream option combinations | 13,393,646 | 130,404.82 | 5,449.64 | 159,964.42 |

Rates use 1,000 actual findings, best of five including process startup. They are bounded measurements, not a general speed claim. The first optional-chain corpus comparison covered both arithmetic settings; the final expanded corpus uses defaults. Underscore corpus parity uses defaults: enabling method-name enforcement over a computed method causes the unmodified Go rule to panic before its name-kind guard. A minimal reproduction and Go stderr are preserved in its owned evidence. Invalid configuration diagnostic parity is outside the decoded-option profile contract.

Mutants all compile and execute cleanly before comparison: ordering-relations-ignored drops unsafe-negation diagnostics with its option enabled; and-left-undefined-ignored loses an optional chain through &&; constructor-property-name-ignored incorrectly exempts this.prototype._bar; block-display-accepted loses Google's block finding. Each is caught on source Node, emitted JavaScript and sanitized native.

Google Font Display is a partial port. Its 27 selector rows include all 16 upstream JSX sources and 11 query/entity cases, producing 4,228 identical bytes against the actual Go rule. Adamic receives extracted attribute data for this bounded comparison. Its complete owned profile builds, but a real JSX link exits 70 on all three Adamic runtimes: `parser slice expected GreaterThanToken, got Identifier at 32`. It cannot yet prove independent JSX parsing, element finding ranges or full corpus parity. No shared file was edited to hide the blocker. Its selector-only rates are documented separately and are not comparable whole-rule rates.

Setup initially failed at cache warming because registration saw a new descriptor before its owned witness existed. After the witnesses were added, `bash cloud/setup.sh` succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 20s, done 20s on 5 processors. The first gofmt call missed the toolchain source; a sourced follow-up formatted the adapters and was pushed separately.

`go run ./cmd/lint-registry` discovers all owned descriptors. `go test` over the registry and the four touched Go rule packages passes. The unmodified Go core rule tests pass in 0.015s and Google Font tests in 0.010s. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1` passes in 0.920s. All outputs went directly to log files. The full repository gate was not run. `docs/parallel-work.md` is absent here; the current CLAUDE.md and docs/lint-registration.md contracts were followed.

## Next queue audit

All completion commits were pushed before this audit. A fresh all-head fetch examined 341 origin refs and 16 distinct claims-directory trees, with origin/main at ef3d907ecdc4c771b016f7d9c52372def057a340. The original 46-rule helper-ready order is the list linked from helpers/REPORT.md into HELPERS.md. Every entry is ported on main or named in an origin claim file.

The inventory's `syntax ready for AST/API adaptation` queue has fourteen entries unclaimed and absent on main. Each already has an executable implementation on an origin branch, so the original instruction to skip existing origin ports applies:

- codex/stage1-lint-batch2: no-script-url, no-unsafe-finally, no-useless-catch, no-useless-concat.
- codex/stage1-lint-batch8: no-octal-escape, no-unexpected-multiline, no-unused-private-class-members, no-useless-constructor, prefer-template, react/forward-ref-uses-ref, react/jsx-no-comment-textnodes, react/no-find-dom-node, react/no-is-mounted, react/no-redundant-should-component-update.

The actual selector modules and all observed origin paths are recorded in wave1-07-next-audit.json. These observations do not recertify those older ports. The separate `syntax waiting on helpers` queue still has work; it is not reported exhausted. No further rule was claimed because no unclaimed, unported candidate remains in the requested helper-ready and ready syntax queues.
