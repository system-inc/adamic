# Full core.ts site census

The exact full-site protocol is the referenced core-sites report's protocol:
LATENT_FULL=1, LATENT_ONLY_FILE naming the complete core.ts, and
LATENT_ASSERT_NO_OUTPUT=1, with the same measurement-only overlay generator.
The ordinary compiler binary is separate and unchanged. All three streams have
zero checker diagnostics. The meter's historical label "measured on a
checker-rejected program" remains unchanged despite those empty diagnostics.

| Measure | Before | After | Original-owner mutant |
| --- | ---: | ---: | ---: |
| Unique non-Boundary sites across all core.ts | 59 | 58 | 59 |
| takeWhile cast, core.ts:2595:16, refusal_scan | Refused | absent | Refused |
| New sites after the adaptation | - | 0 | - |

The exact removed reason is "a cast the runtime can't check", constructed by
internal/lower/cast_proof.go:188. The comparison checks the complete site sets,
not only these totals, and adjusts only the four source lines added inside the
owner. The mutant restores the entire original owner, makes core.ts and all
746 src files byte-identical to before, and restores the complete original
site set. Every raw event, including Boundary records, is retained.

All three takeWhile declarations remain deferred_generic_instantiation with
instantiated=false before, after and in the mutant. Removal of the scan finding
is not a claim that the generic body has lowered. CENSUS.json retains these
unit records and the unchanged raw measurement headers. Raw streams, progress,
timings, tool hash, overlay sources and process-group cleanup records are in
proof.tar.gz. census.py and analyze-census.py reproduce the checks.
