# Slot 03 batch twenty-one

Three helper ports, one per .a file. Go strings cross these APIs as exact readonly byte arrays, preserving invalid UTF-8 as well as multibyte text. Input values must be bytes 0 through 255.

- split_path.a returns basename and parent basename. It takes an explicit lastSeparator dependency returning a Go byte index. The driver uses the previously proven owned batch20/last_separator.a. Go's actual dependency calls are recorded and compared, including their byte arguments and order.
- has_path_segment.a answers whether a nonempty byte sequence is exactly one slash/backslash-delimited path component. It reproduces Go FieldsFunc's separator behavior without interpreting or normalizing the bytes. No regex is replaced or hand matched.
- gap_root_accepts_modifier.a takes the ordered IsColor fields of a valid non-null FunctionalUtilityDescription's Arms. Nil and empty Arms both become an empty array. Other description fields do not affect the Go helper. A nil description is outside the successful-input contract; Go dereferences it.

These do not implement complete rules, rule registration, reports or finding-position conversion. Byte offsets only become UTF-16 at a finding boundary. Temporary Go overlays add observers and exports without changing actual upstream worktree files. No shared harness, registry or compiler file is edited.

The helper gate compares Go, source Node, emitted JavaScript and ASan/UBSan native output, and fourteen compiling semantic mutants. Raw consumer fixtures are captured before upstream checks require a live Tailwind installation. The upstream package gate remains FAILED: eight checks require the absent ahra engine/corpus checkout. Regeneration writes captured helper inputs and coverage metadata, then asserts the actual package exit is zero; it exits nonzero here. Neither that failure nor the absent full corpus is represented as a passing gate. See REPORT.md and evidence/capture.log for the blocker and reproducer.
