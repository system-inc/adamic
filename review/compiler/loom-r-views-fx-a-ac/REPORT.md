Added 21 measured refusal headers with every program body unchanged; standalone a-check is still blocked by three compiler panics.
Commits: source 6c733f966fa1636f1158ea7767a40030f65d2126; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is in the final response.
Commands and outputs: compiled all 54 changed .a files; unchanged pinned Gate.aCheck accepts 51 and rejects three compiler failures; fixture/readers/reader guard/vet checks passed.
Mutants: removed header and wrong reason each rejected by the actual predicate; both original compiler-guard mutants also caught.
Not covered: compiler fixes for the three historical items 136/144 witnesses, full-tree/full-package gates or Loom submission; ready_for_loom is false.

This branch is one commit on dbdf977f5b, containing 6c733f96's change plus headers and their evidence. It has no dependency on a newer main or another compiler change. The carry-over applied without conflicts.

The README was read whole. The exact Gate.aCheck implementation pinned there was fetched from fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py, imported unchanged, and run through its step/spawn/fail harness using real compiler subprocesses. gate-provenance.json and pinned-gate.py.gz retain the exact implementation and hash. This is the README's predicate, not a reimplementation of its acceptance rule. Both scanner and gate build this branch's own compiler. The scan includes all added or modified .a files relative to the requested base, including review evidence, with no exemptions.

| Status | Files |
|---|---:|
| checked (adamic c exits 0) | 30 |
| NotYet | 0 |
| refused | 21 |
| type error | 0 |
| failed: compiler panic, exit 2 | 3 |
| Total | 54 |

TABLE.md is the per-file status/reason/header table; statuses.json contains complete diagnostics and body hashes. headers.json lists each of the 21 exact new first lines. All are `// a-check: refused <measured reason>`; none is a checker-error header. Checked files retain no expected-error header. The 54 program bodies are byte-for-byte equal to source 6c733f96 below any a-check header, including every comment, blank line and final newline. Final SHA-256 values were rechecked after all mutation experiments and control regeneration. No .a file was excluded, renamed or removed, and no new .a program was introduced by this follow-up.

The full actual predicate ran twice on all 54 files, including after mutant restoration. Both runs exit 1 for exactly these files:

| File | Historical inventory item | Actual compiler failure |
|---|---:|---|
| review/compiler/loom-views-fx-a/sources/0ade804a7ad8/misfit_boolean.a | 144 | panic: compiler bug: narrowed scalar union view requires a checked conversion |
| review/compiler/loom-views-fx-a/sources/a705d6d0d5df/misfit_number.a | 144 | panic: compiler bug: narrowed scalar union view requires a checked conversion |
| review/compiler/loom-views-fx-a/sources/b057160c0f66/p51.a | 136 | panic: compiler bug: incomplete checked field contract |

These failures are present with their original 6c733f96 program bodies. A compiler panic is outside the four allowed outcomes and cannot truthfully receive a refused or type-error header. The unchanged predicate accepts all other 51 files, verified by a separate real-compiler run. These three files remain in the full final scan and in the delivered change. summary.json explicitly sets ready_for_loom=false. A standalone green result on this fixed base needs those compiler failures resolved; a header-only pass cannot provide it. The existing source inventory identifies them as the historical B-unit items 136 and 144, outside A's 132/142 compiler changes. No additional compiler change was imported or invented for this header unit.

The scan driver initially stopped at the first unexpected compiler panic. That attempt is retained as scan-initial.log.gz and statuses-initial.json. It was then extended to record failed outcomes and finish all 54. The predicate driver exits zero only after asserting the known full-scan failure, accepted-subset success, both intended mutant failures and restoration. Its process exit is not a claim that the full a-check passed: predicate-final.json retains a-check exit 1.

Header proof: key.a was changed on disk to remove its first-line header, then independently to replace its reason with an intentionally wrong reason. Each was compiled by the real branch compiler through unchanged Gate.aCheck. The compiler still refuses the same program; the predicate rejects the missing expectation and the mismatched reason respectively. Each mutant's a-check exit is 1. The original header is restored in finally, and the final full scan has only the same three unrelated panic failures. header-mutants.json, mutant-*.json and the compressed patches record the catchers. Both source compiler-guard mutants from 6c733f96 were also rerun and caught by TestLoomViewKeyRead and TestLoomViewUnsupportedRead with got <nil>; both source files were restored.

Position and reader audit: exact file paths were searched in *_test.go and testdata, and the dynamic loom_views_a fixture reader was inspected. Its tests require the reason, fix and a located .a diagnostic, with no pinned numeric line or column. No Go test or testdata position expectation needed editing. The current replay's diagnostic observation file review/compiler/loom-r-views-fx-a/entrypoints.json was refreshed for the shifted lines. Older baseline reports remain historical observations of their recorded compiler/source snapshots, not current position expectations. The replay generator now strips a-check metadata when deriving its two positive controls, preventing a refusal header from being copied into a clean program. Its regenerated control bodies are exactly equal to 6c733f96. These are the only non-header adjustments to existing live replay code/data.

All thirteen TestLoomView leaves pass with headers in place and continue to pin source Node before checking refusals. The CLI replay also confirms all eleven probes refuse with the intended location/reason/fix in both backends; its two positive controls agree with source Node in JavaScript, release native and sanitized native, with empty stderr and exit zero. The 52-package inverse-import closure from the prior same-base delivery was rerun with the focused fixture/flow selection; every package passed, with unmatched packages compiled only. Flow coverage and relevant program units passed too. TestCallTargetReaders passed in 4.32 seconds; go vet ./internal/lower passed with empty output. The final focused fixture rerun passed in 0.530 seconds. Full output is captured to files, never piped. No full package suite or full gate was run.

| Existing fixture leaf | Reader pass seconds | Final fixture pass seconds |
|---|---:|---:|
| TestLoomViewKeyFix | 0.29 | 0.22 |
| TestLoomViewKeyRead | 0.28 | 0.20 |
| TestLoomViewOriginalNativeP21 | 0.38 | 0.20 |
| TestLoomViewOriginalNativeP37 | 0.40 | 0.22 |
| TestLoomViewOriginalViewsP01 | 0.37 | 0.13 |
| TestLoomViewOriginalViewsP05 | 0.47 | 0.14 |
| TestLoomViewOriginalViewsP06 | 0.30 | 0.21 |
| TestLoomViewOriginalViewsP19 | 0.33 | 0.15 |
| TestLoomViewOriginalViewsP72 | 0.29 | 0.22 |
| TestLoomViewUnionKeyRead | 0.34 | 0.19 |
| TestLoomViewUnsupportedCallableRead | 0.26 | 0.12 |
| TestLoomViewUnsupportedRead | 0.21 | 0.15 |
| TestLoomViewUnsupportedUnreadControl | 0.44 | 0.27 |

No compiler Go file or Go test differs from 6c733f96. No runtime C line changes. No oracle-counted source or counts row changes; counts.md remains blob bdab0f5b96a2080ca1220efe4d18de2a693fd62d. No counts regeneration was needed for headers confined to lower/testdata and review files. The prior same-base source change's full Linux regeneration remains recorded in its report.

Setup succeeded using GOPROXY=https://proxy.golang.org|direct and the printed env.sh. nproc=5, with a four-CPU cgroup quota. Timing lines:

```
submodule cache: off
v24.19.0
setup: node ready (0.026s)
go version go1.27.1 linux/amd64
setup: go ready (0.031s)
setup: submodules ready (0.052s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.010s
setup: markdown dependencies ready (0.084s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.219s)
setup: shared cache ready (0.596s)
setup: go build ready (122.436s)
setup: test binaries deferred (use --warm-tests) (122.511s)
setup: build cache warm (122.513s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/136 765 load-after=6.45 2.13 0.76 1/147 1788
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (122.539s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.wmINd3
```

Setup exhausted the inherited root filesystem's available user blocks. Initial per-file cache relocations could move bytes but could not recreate symlinks while the filesystem was still below its reserve threshold. A batch move of newly materialized immutable cache data to temporary storage restored available space; every moved cache path was then restored as a symlink to the preserved bytes. cache-relocation.json records the data. The branch compiler was built in /tmp and all checks then completed. The reader guard used temporary product storage for its large export copies. No compiler, program body, or cache-key logic was changed for this workaround.

Conservative scope choice: retain all review witnesses and honest failed outcomes. There is no supported a-check header for a compiler panic. The supplied base and 6c733f96-only compiler scope exclude silently pulling in the separate items 136/144 fixes. This delivery is reviewable but not ready for standalone Loom submission. No ahra CLI, PR, main push, force push or Loom submission was made.
