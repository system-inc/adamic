# Finite computed fields, 2026-10-08 UTC

Fixture 25 at 724:111 reads wildcardMatchers[usage] as its object-pattern default. The checker proves a finite string key and fields all of type WildcardMatcher. Lower selection through an ordinary helper that captures receiver and key in source order. Candidate fields must be required ordinary data fields with identical checker types and a supported stored representation. Preserve inherited prototype refusal diagnostics, erased method origin checks, and conservative accessor/class/optional-field limits. No checker contextual type changes.

The finite_property_index.a probe agrees with Node in both backends, with default selection at all three keys, explicit override, and receiver/key effects once in order. Node stdout: owned:files, owned:dirs, owned:exclude, override, owned:dirs, rk on successive lines. Counted alloc/free 10/10, retain/release 15/20, peak 9, regions 0. The focused uncached oracle (including previous fresh-map, defaults and positioned-search regressions) passed, 0.424s.

Every key comparison returns that field. A stale narrowed key outside the checker's union reaches an explicit panic. The separate invalid-key source mutates a captured key during receiver evaluation: Node reports TypeError at the missing field; both Adamic backends check the violation instead of selecting an arbitrary field. This negative witness lives in lower/testdata and is not a successful counted oracle fixture.

Executed selection mutant returns the first field in every matching branch: both backends exit 0 with wrong stdout and the oracle fails. Executed key-check mutant replaces the final panic with a field return: both backends incorrectly succeed and TestFinitePropertyInvalidKeyChecked fails. Logs /tmp/finite-property-selection-mutant.log and /tmp/finite-property-key-mutant.log. Both mutants restored.

Full counts regeneration remains blocked by the previously reported regexp_tree project-root attribution gap. No complete fixture 25 pass claimed; rerun the pinned scratch merge after landing this stop.
