Rebased wave 11 onto lint integration b4691483, including main c7991b90 and the inherited legacy registry migration.
Green rebased tip 4ccdc6a5a pushed only to codex/typeaware-wave-11; evidence commit follows.
All eight owned suites, full bridge tests, filtered compiler oracle and package vet PASS.
All 46 worker mutant observations revalidated through independent byte comparison and ownership checks.
No unclaimed rules: 644 origin refs, 629 distinct trees, 34 reservation documents; no new claims.

Complete commands, exits and logs are adjacent. Seventh process 95.682s,
eighth 112.012s, bridge 76.065s, compiler 2.107s. Controls and both frozen
corpora match including findings/fixes/suggestions under normal and sanitizer
builds. Released handles and ownership mutants pass their requirements.
Evidence writing initially failed for disk space after checks; named obsolete
worker scratch archives were removed, retaining source and logs. No shared
harness/generator or protected compiler source changed. Inherited runtime,
allocator-helper and registry migration changes were retained. Main and area
heads were verified unchanged at push.

Full gate and its 17 required external-input checks were not run. Standalone
worker validation does not claim shared name-based registry compatibility
for older numeric metadata. Shared emitted-JavaScript lint execution and
nondefault-option limits remain uncovered; four analysis claims remain parked.
Original quiet native/Go times remain repository 0.9227/0.1380s and compiler
7.0781/0.3421s on c01907a7; current concurrent timing observations are in logs.
