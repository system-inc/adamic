# Node global console identity, 2026-10-08 UTC

At fixture 25:1296:5 the global console has only the pinned @types/node declaration, not the Adamic prelude variable. Recognize that declaration as the same supported intrinsic, keeping single-string arguments and log/error stream selection. User console objects still lower through their own methods.

stage3/fixtures/host/console_identity.a first reproduced the reading-console stop, then matched Node in both backends, sanitized/release and leaks, uncached 0.980s. Cover Node-only declarations, an owned string, stdout/stderr and a local shadowed console mutating a captured value. Counts 6 allocations/frees, 6 retains, 11 releases, peak 3, regions 0. The spelling-only mutant recognizes the local console too; both backends exit normally with an extra newline and fail Node stdout (/tmp/node-console-identity-mutant.log). Restored.

Pinned fixture 25 next stops at 1175:17 on a boxed union captured by a closure (/tmp/node-console-host25.log). No whole-fixture pass, new nullable kind, or whole-scratch gate is claimed. The Node declaration loader is a host dependency absent from the base branch, so this fix lives in the scratch branch.
