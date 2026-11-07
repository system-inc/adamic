# Full-apply checkpoint

Base: origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8.

| Check | Result |
| --- | --- |
| Whole-lane measurement | 398 units; 860 integer calls; 0 excluded-class calls; 0 counterexamples |
| Whole-lane control and reconciliation | Pass; exact test multiset and baseline diff |
| Baseline fresh full apply with product packaging | Exit 0; 81,504 inputs match built baseline |
| Main plus arrayof full apply with product packaging | Exit 0 |
| Owner and no-write contract | 11 checks pass |
| Before types | 79 roots; exit 0; 0 diagnostics |
| Before full-core census | Exit 0; F2 present at core.ts:1339:12 |
| Paired replay, after types, JavaScript, declarations, census and mutant | Pending |

The failed initial baseline product-packaging attempt and layout preflights are
retained separately. They are excluded from accepted measurement and apply
results. The frozen adapter and owner pins are unchanged by evidence packaging.
