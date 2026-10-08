# ecmascript/classmembers package claim

Branch: lint-helpers/ecmascript-classmembers. Base: origin/area/stage1-lint.
Triage d7ab0bc4, rank 20: no retained port, five needed symbols; build fresh.
Higher eligible packages are remotely claimed, excluded tonight, or runtime-regex
blocked. The regexpattern race was yielded to the earlier df0e6952 claim.
All origin implementation paths and all helper-branch claims were checked.
No classmembers port or claim existed at this snapshot.

Helpers: ForEachDuplicate, IsAccessorKind, IsOverloadSignature, KeyOf, MemberName.
Consumers: no-dupe-class-members and @typescript-eslint/no-dupe-class-members.
Completion forecast: zero alone; two additional with preceding packages, 129
cumulative. These counts are conditional, not completed rule parity claims.

Audit ecmascript/property.NameTagged before dependent work. That package is being
landed by another worker tonight: use its shared implementation when available,
otherwise stop the dependent helper rather than duplicate it privately.
No runtime regex compiler is used by this package.

Claim precedes code; fetch again and yield to an earlier competing reservation.
Require Go consumer-capture agreement on Node, emitted JavaScript and sanitized
native; one compiling, running mutant per helper on Node and native. Prove one
rule when its shared prerequisite closure is available; full helpers/lint gates
use every required input before the finished-unit push.

Post-push claim scan: sole classmembers reservation, 2a7d5c0135769fe9bed15cd32edb0c0e42134ca2 at 2026-10-08T12:21:36Z.
Status: all five helpers validated against live Go consumer captures. Shared property.NameTagged was built at ecfd7ffc1e4d37c8b1933c8b339913e2cd3c9a59 under explicit follow-up authorization. Core no-dupe-class-members is the proof rule; the TypeScript extension is now helper-ready and remains with its existing claimant.
