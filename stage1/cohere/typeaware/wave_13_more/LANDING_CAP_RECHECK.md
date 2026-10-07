Built: verified the existing nine implementations are already rebased onto current main; eight remain green and no-object-constructor is partially blocked. No new claims or source changes.
Commits: fetched main e8ba3d5d and own remote 6e2bfe7e; the latter contains the full landing gate evidence and already has main as an ancestor.
Commands and outputs: explicit fetch of main and codex/typeaware-wave-13 succeeded; replayed all three blocked controls against production Go, normal native and ASan native; Go exits 0 with 699/699/611 bytes, native exits 70 with zero output.
Mutants: none repeated in this evidence-only recheck; unchanged current-main evidence in LANDING_REPORT.md records nine rule and seven fact byte catches, seven question guard failures, seven foundation catches and the released-registry catch.
Not covered: three shared-parser inputs remain blocked; full repository Go gate, emitted-JavaScript parity and whole-source cohere CLI lint of .a remain unclaimed. No fully green whole-branch result is claimed.

The only branch pushed by this unit is codex/typeaware-wave-13. It is already
based on current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965 and pushed
at 6e2bfe7e3ca918e382da00188da04a7a74661f8c before this evidence update. No push
to main or area/ was made. No rebase is needed while that main tip is unchanged.
Plain git fetch populated only FETCH_HEAD because of local branch tracking;
explicit fetch refspecs refreshed both remote-tracking refs and confirmed the
published SHA. Existing executable rule, bridge, parser and compiler sources
match the previously tested landing state.

Three exact controls were replayed, with each process stdout/stderr saved to a
file. Production Go and both native binaries are the final-main binaries from
the prior landing gate, so no stale compiler or new build is involved:

| Case | Source shape | Normal and sanitized native refusal |
| --- | --- | --- |
| 53d1c0a9ffc2fefa | `<foo />` then Object() in .ts | expected GreaterThanToken, got SlashToken at 5 |
| aff6ced2ca61fe89 | `<foo></foo>` then Object() in .ts | expected GreaterThanToken, got Identifier at 7 |
| defcb4c4ce6a2921 | yield label and break yield | expected semicolon at 37 |

Each Go process exits 0; each native process exits 70, emits no stdout, and has
no sanitizer report. These are unchanged shared-parser refusals, not missing
rule logic or suggestion serialization. Ahra said: "Keep your changes inside
your own rule directories" and "If anything else blocks you, say exactly what
it is and stop, rather than editing shared files." The parser lies outside this
unit's rule directories, so no shared parser edit or fixture rewriting was made.
No additional rule claim is permitted while the claim is partial.

The prior current-main gates remain applicable because no executable source,
submodule, toolchain or main SHA changed: original trio PASS 418.694s, bridge
PASS 318.896s, expanded Node oracle PASS 17.193s, go vet PASS; normal and sanitized
Nexus controls 129/129 equal and supported core controls 430/433 equal. The prior
isolated compiler medians remain 8.503342s native / 2.363765s Go for original,
3.034379s / 0.547875s for Nexus and 3.478860s / 0.421988s for core. No new timing
claim is made. Existing toolchain reused; setup previously reported 0s for Go,
clang, Node and submodules, 200s cache warm, 200s total. nproc remains 5.

validation/landing-cap-recheck retains the explicit fetch log, exact observed
refs, per-process exits, complete output streams, parser messages and hashes.
No test output was piped, no PR was opened, and no other branch was pushed.
