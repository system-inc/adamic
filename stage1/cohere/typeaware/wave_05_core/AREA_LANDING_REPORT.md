Built: nine completed wave-05 ports rebased onto the shared-harness area; JSX blocker closed, React analysis claims remain parked.
Commits: tested rebased tip 41250962161886d8ad4fa754221ead78f8e2ddaa on area 7481e0324e34a2537aafa9db7eeacda50405611b and current main 39638d9e; enclosing commit records evidence.
Commands and outputs: owned rule controls/corpora, sanitizers, releases, bridge, filtered uncached Node, registry, metadata and vet PASS.
Mutants: nine rule verdicts and CFG corruption caught only by Go bytes; release-registry, listener and integrated-registry checks pass.
Not covered: React rule verdict parity and native analysis, full repository gate, full option matrices and migration of private ports into shared registry.

Ahra requested rebasing onto origin/area/stage1-lint. The fetched area had advanced beyond the named 50a5f105 to 7481e032, which includes harness 41eb6eab2 and current main 39638d9e278d38bb5aeae887f46d55a70e47aaad. All 25 wave commits applied without conflicts. Shared harness, parser and integration changes are retained; none is reverted. No shared implementation file is edited in this landing. Only codex/typeaware-wave-05 is published, with an exact lease on previous owned tip 24561b6ed065fb8773879380abd8f7002c10c9dd after rebase.

All nine completed ports retain their private type-aware drivers and independent production Go oracles. The seven-argument shared finding interface and wire compatibility require no rule change. The three latest rule manifests use validated Go AST kind names and callbacks receive their cached node. Integrated registry package tests pass, but shared-registry registration of these older private ports is not asserted.

Revalidation commands are the prior AREA/LANDING commands using area log prefixes and /workspace/wave-05-area-first-artifacts. Every toolchain shell sources /workspace/adamic-tools/env.sh. The original-three and timer/output validators explicitly receive ADAMIC_TYPESCRIPT_SOURCE and both frozen corpus manifests. Test output goes directly to logs:

- TestWave05AgreementAndMutants: original three rule controls, both corpora normally and ASan/UBSan/LSan, three verdict mutants and release-registry deletion mutant.
- WAVE05_OUTPUT_SKIP_BENCH=1 validate_output.py: both output rule controls/modules, DOM/Node timer regressions, verdict/CFG mutants, releases and both sanitized corpora.
- WAVE05_NEXT_SKIP_BENCH=1 validate.py: timer controls, verdict mutant, release, historical constructor-gap probe and both sanitized corpora.
- wave_05_core/validate.py: 106 parse-valid controls, 76 findings and complete suggestions, three verdict mutants, two releases, both corpora normally and sanitized.
- go test ./bridge/tsgo/... -count=1 -timeout=15m: bridge PASS 100.648s, checker PASS 0.498s.
- ADAMIC_GATE_UNCACHED=1 filtered internal/oracle: PASS 2.303s, native 0 hits/28 misses, Node 0 hits/19 misses, including the one-byte comparison mutant.
- go test ./stage1/cohere/lint/registry -count=1 -timeout=5m: PASS 0.077s, integrated registry validation/mutations.
- Both metadata scripts and go vet ./bridge/tsgo/...: PASS.

## Updated React dependency

The old JSX parser blob bc0ee72 has been replaced by 62ab6514da477dfc54f477b6a968735e24901b58 on this requested area base. Rebuilding the retained jsx_probe.a and running globals.tsx, immutability.tsx and derived.tsx now exits zero with `jsx nodes 2` for each, with empty stderr. The non-JSX control exits zero with `jsx nodes 0`. The old refusal verifier is historical evidence, not a current expected failure. This demonstrates parsing only, not React rule parity.

The claims remain PARKED for missing native analysis: globals needs React component/hook compilation-root and capture/reference-write classification, including the existing Go helpers IsComponentOrHookLike, IsReachableRootPosition and WritesToBinding; immutability needs HIR/SSA/capture and frozen-value analysis; no-deriving-state-in-effects needs HIR/SSA/capture and effect taint/dependency analysis. The HIR/SSA/capture dependency remains #dnv6f2c. JSX is no longer named as an outstanding blocker on this branch.

The refreshed inventory scans 597 origin refs and 33 unique Markdown claim blobs. All 197 ranked checker-dependent names are covered by 172 claims or 25 ranked baseline ports; no unclaimed rule remains. No new reservations are made. Inventory and exact logs are in evidence/area-*. No new uncontended performance timing is claimed because these checks overlap. Prior measurements remain historical. Setup remains the successful 82-second run, nproc 5. Full repository gate, upstream option matrices, shared emitted-JavaScript rule comparison and automatic application of suggestions remain outside this revalidation.

Final observations: original-three oracle PASS 160.312s; output and timer validators PASS; core validator PASS; bridge, Node, integrated registry, metadata and vet PASS. Current main remained 39638d9e immediately before publication. The remote area advanced during checks to d65a8f931c98655936ae04c6899f38f14862b73e; this report certifies fetched area 7481e032, which includes the requested harness, and makes no claim about later area changes.
