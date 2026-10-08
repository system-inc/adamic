# Compiler review

The compiler reviewer approved the staged resolutions before each of the three merges: stable emitter identities with area graph/async/ownership behavior retained; stable module ownership and selective declarations; outlined main with async startup/drain preserved.

Final review of the build-path diff found no code blockers. It covered generated-source and shipping guards, CPU/environment job precedence, runtime worker completion, archive ordering and atomic cache publication, async frame typedef classification, async MainModules metadata, and the generated abandonment harness's stable-symbol discovery. The harness preserves all four cancel/exit control/mutant checks and rejects absent or ambiguous symbols.

This approval concerns code review. Landing also requires the complete uncached verification and raw recorded program-output comparison documented in SPLIT_BUILD.md.

Subsequent oracle findings received compiler review: the area-added initializer type whitelist and narrow handling of external adamic_parallel_map/adamic_parallel_map_move prototypes. The latter requires a nonstatic prototype with no body or initializer and preserves runtime ABI names.

Final compiler review approved landing on the 4,113 byte-identical normal executions and complete 9,172-key/exit-code coverage, requiring explicit disclosure that 31 raw harness/mutant records differ. The committed cold timing is 5.357 seconds; 4.717 seconds was an earlier development measurement. Remaining native checks must finish before push.
