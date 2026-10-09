Built: initial callable metadata, lazy callable descriptors, aggregate payload demand and checked callable read wiring.
Commits: first V4 subset on 2e10854ce0be8d4439c1664e1fdfe13523f4341f; see Git history for its SHA.
Commands/results: focused lower 0.600s, native 13.279s and JavaScript 0.714s, all passed.
Mutants: pending; this checkpoint certifies compilation and focused controls only.
Uncovered: factory, Set domains, witness imports, ruled skips, counts and complete comparison sweep.

This is the compiling subset explicitly requested for the first push, not a completed V4 slice.
Source is a callable-only net cut from lane 5 at 926a1d39. Shared historical merges are excluded.
The callable scaffolding originates in 8346b28c and 9fc5c01a; deferred descriptor and read
helpers in be6e4f33 and 9c4d0904; fixed scalar contracts in 37fa06d3; metadata and
boxing in cb6adb25; aggregate payload demand in 3e1a7f6f, ebb1f8a6 and 164dac2c.
These are source provenance, not whole cherry-picks. Marker helpers from b6e53fb4
are present, but the two admission hunks excluded by judgment 10 are not applied.

Resolutions: preserve the V3 plain/count-aware typed closure and method convention,
including actual argument count, receiver flags, optional/rest slot packing and
argument-index guards. Producer identity uses the existing unionClosureCodeIdentity.
Do not fabricate implementation metadata. The absent ClosureOperands historical
helper is replaced at the aggregate-demand access by this base's CallClosure.Closure.
Optional payload field names are recorded without admitting wider writes.
The callable runtime consumes V3's existing readiness-aware union snapshot instead
of importing historical broader optional-field read/write implementation.
Judgment 10 retains stored never-rest call and marker result-erasure refusals.
Judgment 14 preserves base null/undefined representation. Judgment 15 retains
unproven intersection masks. Remaining ruled decisions apply in later V4 imports.

Commands are recorded directly in hooks-fourth.log. Setup details remain in setup.log.
No whole package comparison or language ruling is claimed at this checkpoint.
