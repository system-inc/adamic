# 99x verification in progress

The initial missing-source stop is historical. On continuation, cloud/setup.sh
completed, npm ci installed stage3/api's pinned API, and the pinned upstream
source was cloned into /workspace/typescript-source.

The baseline command ran from an isolated worktree of main 3e395f9e5, excluding
this branch's then evidence-only adapter directory:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/bind-before
```

Exit 0. Full log: baseline-apply.log.gz. Pin generation reports 46 sites,
15 owners, zero refused core calls and six removable import specifiers.
The pristine scan excludes both binder.ts local helper calls.

| Check | Current result |
| --- | --- |
| 1 baseline full apply | Pass |
| 1 after full apply and ordered replay | Pending |
| Owner and replay contract | Pending |
| 2 all 79 roots, Adamic types | Pending |
| 3 nine emitted JavaScript files | Pending |
| 4 every oracle shard | Pending |
| 5 restored-site negative witness | Pending |
| 6 ordinary and latent census | Pending |

CW051's six recorded units are checker.ts:54277:1, watch.ts:673:1,
watchPublic.ts:123:1, executeCommandLine.ts:812:1,
executeCommandLine.ts:1011:1 and executeCommandLine.ts:1034:1.
Their before/after measurements are pending.
