# Byte-preserving regex flags and literal helpers

Use parseRegexFlags(bytes) from parse_regex_flags.a, regexFlagsUV(flags) from regex_flags_uv.a, and patternAndFlags(bytes) from pattern_and_flags.a. Bytes are the original UTF-8 Go string representation, integer values 0..255; results preserve invalid UTF-8. RegexFlags is declared in flags.a. Split results own mutable byte-array copies. These functions inspect flags and delimiters and do not match regular expressions, validate flags or construct findings.

Go behavior: u/v independently present anywhere; UV uses logical OR; a literal needs length at least two and an initial slash, splits at its last later slash, and has an empty flags slice when no closer exists. Empty/nonliteral input returns both arrays empty. Consumer adapters retain byte offsets until building finding positions.

See REPORT.md for the actual Go, source Node, emitted-JavaScript and sanitized native oracle, all twelve compiling mutants and limits. Run the reported go test command after sourcing the setup environment. Frozen readiness consumers for each helper: no-control-regex, no-regex-spaces, no-useless-escape; nine prerequisite occurrences removed, no final blocker completed. Per-mode evidence covers every inventory consumer, eight for flags/UV and nine for literal splitting.
