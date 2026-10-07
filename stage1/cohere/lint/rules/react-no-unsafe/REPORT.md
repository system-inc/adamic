Built `react/no-unsafe` in Adamic, preserving component ownership, alias defaults and key spans.
Validation: all 51 captured upstream cases passed byte for byte with Go on Node, emitted JS and sanitized native; shared with sort-vars, 122 cases / 36,855 bytes.
Mutant: unsafe_component_detection_disabled compiled and ran, then failed only output comparison on all three runtimes.
Findings/s, native / Node / Go: 874.76 / 902.54 / 4518.58, 77 compiler files plus 1,000 positive findings, best of five.
Limits: unmodified shared registration/profile gate still needs codex/lint-harness-dot-a; no shared source was edited.

Claim 7915fd3 preceded source. The final witness test passed 1,888 bytes. Initial corpus comparison covered 223 files / 669 cases / 37,137,620 bytes across all three original rules. Final corpus evidence is stored with sort-vars. Go callbacks, messages and judgments are unchanged in the external oracle. Captured decoded options are unmarshaled into the original Go type.

An initial mutation removed the component guard but did not change the positive-only witness and survived. It is retained in the evidence. The final mutation disables actual component detection, loses the positive finding, compiles/runs successfully and is caught on all three runtimes. Nothing killed only by compilation is credited.

Tests use the already-owned Tailwind scratch compatibility tools, never modifications to shared repository files. See sort-vars/validation_test.go.txt and the adjacent evidence for commands and output. The original raw setup failed shared profile compilation; scratch setup succeeded in 17s on 5 processors (Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 17s).
