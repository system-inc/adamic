# Triple slash reference

Preserves TypeScript filename gating, decoded option defaults, leading-comment ranges, attribute trimming, top-level import matching and the last repeated types directive.

The implementation uses `.ts` under Ahra's explicit fallback until the shared harness supports `.a`; the independent driver is `.a`. Each directory owns its descriptor, unchanged Go rule adapter, raw witness and semantic mutant. No shared source was edited.

The independent comparison replays 70 upstream cases for this rule, all cases returned by successful Go test runs, with no exclusions. It also covers TypeScript's compiler sources and the current stage1 sources on source Node, emitted JavaScript and sanitized native. It compares rendered findings, message ids, diagnostic ranges, repair text, actual repair ranges and final fixed source against unmodified Go rules. See [the batch report](../typescript-prefer-function-type/validation/REPORT.md) for exact coverage, commands, evidence and measured throughput.

Keeping the first repeated directive instead of the last changes a finding span without affecting compilation. The full comparison rejected that mutant on all three runtimes with exit zero and empty stderr.

The default lint package still fails to compile because `profile_test.go:32` ranges over the `portFiles` function. The shared serializer rejects fixes whose spans differ from diagnostic spans, and `Linter.fixed()` applies replacements to diagnostic spans rather than `editStart`/`editEnd`. The two fixing rules therefore mark edits `range-fix`, retain their actual ranges, and are validated by the rule-owned driver. The default fixer skips these edits safely. Adopting them in the shared driver requires its owner to preserve both ranges, apply actual fix ranges and handle this marker. Triple-slash-reference does not offer fixes, so only the package compilation issue blocks its default integration. These are independently validated candidates pending shared infrastructure integration.

During final validation the shared harness branch became available at `2650ad59`. It supplies `.a`, profile and actual-edit serialization support on a newer base. This unit did not merge that broader branch or edit shared files. The blockers above describe this checkout; integration should use the published harness and change the two owned `range-fix` markers to `fix` once its actual-range fixer is present. No default-harness pass after that integration is claimed.
