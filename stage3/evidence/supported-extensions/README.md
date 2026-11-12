# STOP: main advanced again in the same owner

No implementation commit was pushed. The reviewable uncommitted draft is preserved in [draft.patch](draft.patch). Completed measurements use main 87028b0889b4431a08c31ca70bc294fa0fd746f5. Main then advanced to abebb00bf6312fb92798095515fbee86b8ed78e3, whose lowering change explicitly targets getSupportedExtensions. Those completed compiler proofs do not certify the new tip.

On the measured base: all ten emitted JavaScript outputs and both public declarations are byte-identical; stock tsc CLI and adamic types report zero over both 79-root sets; both full applies, the adapter contract, zero-edit replay and generate_shards --check pass. Each full census completes 933 units. Exactly the 10005:158 cast refusal disappears, nothing is added, and the restoring mutant reproduces baseline findings exactly.

The keyed lane product passes. The whole lane is **incomplete**: all 398 shards were registered, 84 passed, and no failures were reported before controlled termination. The Go child's actual signal exit is -15 (shell exit 143). This is not a passing whole-lane verdict. See [summary.json](summary.json), [main-advance.json](main-advance.json), and the raw logs.

# Supported extensions cast evidence

Implementation branch: typescript/supported-extensions-cast. Measurements are kept only on typescript/supported-extensions-evidence. The final summary records the implementation commit and completed verdicts.

The direct readonly Extension[] to readonly string[] upcast probe lowers successfully. Its restoring string-to-Extension control is refused. Both profiles are retained, including an initial console-signature mistake and its corrected reruns.

All 79 compiler roots are checked with stock TypeScript 6.0.3 under strict, noUncheckedIndexedAccess, exactOptionalPropertyTypes, ES2024, ESNext, Bundler and the existing pinned Node ambient declarations. Both root sets are also checked by adamic types. The complete source comparison establishes that the measured candidate equals fresh full apply output byte for byte.

The output comparison checks all ten built/local JavaScript outputs, including the test harness, and both public declarations (typescript.d.ts and tsserverlibrary.d.ts). The adapter's own transpile checks run with and without comments. Its contract checks whole-owner drift refusal, both line endings, and zero-write replay with bytes and mtime preserved.

The full-mode census uses LATENT_FULL=1, LATENT_ONLY_FILE set to each tree's utilities.ts, LATENT_ASSERT_NO_OUTPUT=1, and per-unit timings. A complete relation_report trailer is mandatory. Comparison deduplicates actual sites by normalized source location, kind and reason, excludes Boundary and SkippedDependency events, and requires exactly the original 10005:158 refusal to disappear, no added sites, and exact restoration in the mutant. Phase-level deltas are retained too.

Main advanced from 02ebf2b154495843ba2b1c51de63ed8bd9a8e3b9 to 87028b0889b4431a08c31ca70bc294fa0fd746f5 during verification. The implementation was rebased. Prior-tip censuses were terminated, kept as incomplete observations, and replaced using rebuilt compiler and census binaries. provenance.json proves identical stage3 adaptation/build inputs across that advance and records compiler hashes.

run.py records and reports each child's own exit code, kills its process group and reaps its direct child, and limits each invocation to 1700 seconds. Negative signal codes are retained in JSON; shell exits use 128 plus the signal number. Physical Go evidence filenames end in .go.txt. The original guarded overlay generator is changed only to give its scratch Go filenames that suffix. Production compiler sources are not modified by the adaptation.

The lane uses its own unmodified keyed product recipe, then the whole stage3/lane Go package, including every one of the 398 manifest shards. generate_shards --check is also run. Raw logs, statuses, census records, scripts and SHA-256 manifest accompany the final summary. Historical failed setup attempts and deliberately refused controls are not promoted to passing proofs.
