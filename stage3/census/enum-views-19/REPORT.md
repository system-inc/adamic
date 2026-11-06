# Enum views: nineteen-site B2 measurement

Target: 9bdbe75ede31d9b18e49ae3a79843eeaa89c63d4, change chg_bzxjsxe1,
from typescript/enum-views-readonly-2. This is a fresh worktree and a report-only
branch, typescript/enum-views-19. Adaptation 40 is unchanged.

The two earlier columns come from the prior retained measurement: main
63e14a0b2b61cf251ad74292e42d5fd09af804b3 had writable-record Refused at
invariance.go:673; adaptation 40 alone had namespace-value NotYet at
module_namespace.go:27. All nineteen rule IDs are retained: eighteen casts in
debug.ts plus rule 91 in emitter.ts, matching SITE-STOPS.md.

B2's four Library changes end at ff09c4494. Their parent is
7b66fad4dbece492846a10f878417ab9ecd1e36b, two main commits later than the
retained main baseline. The ordinary comparison measures the complete target
against that explicitly pinned earlier main, without attributing every change
exclusively to Library.

## Completed pieces

- Setup: the requested commit is checked out, pinned dependencies match, and
  the ordinary compiler, existing full latent census, and replay worker build.
  Raw provenance, tool hashes, build logs and reference records are in
  01-setup.tar.gz.

## Pending measurements

The target's full apply is running. Next pieces measure debug.ts's full latent
units, emitter.ts's nineteenth owning unit, and the ordinary 79-root census.
Each run is bounded below 25 minutes and each completed piece is committed and
plainly pushed. Cast scan admission and actual owning-unit lowering outcomes
will be reported separately, including earlier boundaries where present.

All raw records are packed in the numbered tar.gz files. SHA256SUMS hashes the
archives; each archive also contains a SHA256SUMS for its raw member files.
