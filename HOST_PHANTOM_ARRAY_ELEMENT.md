Closed host fixture overload refusal at 678:1 under existing phantom-brand rules.
Commits: scratch merge cadf3598 contains feature 52882bdf and host c97402ba; this proof follows it.
Checks: phantom overload lower controls .210s; Node/native/JavaScript oracle 23.934s; counted witness .187s.
Mutant: replacing bidirectional primitive compatibility with representation equality loses the branded literal refusal; TestPhantomArrayElementOverloadLiteralRefused catches got <nil>.
Not covered: full scratch lower gate fails existing predicate-marker test and virtual-call target panic; fixture 25 next stops at 724:5, a destructured parameter with a default; no new null sentinel kind reached.

Date: 2026-10-08 UTC. Pinned fixture 25 blob remains bd24f7bd2ed712f6d07455130ab8bbe76c82b743.

The host already admits void phantom scalar and array brands. Its overload
result proof compared branded primitive array elements using raw assignability,
which rejected string[] -> Path[] despite identical accepted phantom views.
The extension compares primitive elements bidirectionally under the existing
phantom relation, then retains array readonly identity, ownership and mutable
slot checks. A branded literal element remains refused. The witness constructs
owned strings, returns branded components, pushes another branded element,
and matches Node in both backends. Counts: allocations/frees 8/8, retains 5,
releases 11, peak 5, no regions or graph operations.

The nullable dependency had a second sparse constructor. This scratch uses the
host's established array_holes.c constructor and retains the dependency's
presence helpers. Other numeric typed-array and sparse behavior were reconciled
around the host implementations. This scratch is a fixture proof, not a claim
that all combined host/dependency tests pass. The actual full lower gate logs
are host-phantom-overload-lower*.log in /tmp. The focused proof, literal mutant,
backend and next-stop logs are host-phantom-overload-{controls,literal-mutant,
backends2,counts}.log and host-phantom-next-stop.log.
