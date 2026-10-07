Built: tailwind.FindEntryPoint in one .a helper file, preserving candidate priority, Linux filepath.Join behavior, callback probes and short-circuiting.
Commits: claim 48292211 pushed before implementation; this report accompanies the implementation commit.
Commands and outputs: original Go consumer suites PASS, 154 actual helper calls from all six consumers; 922 cases and 130,833 bytes match Go on source Node, emitted JavaScript and sanitized native; owned package test and vet PASS.
Mutant: entry_point_priority_reversed compiles and runs cleanly, then only byte comparisons catch it on all three backends.
Not covered: Windows filepath semantics, complete Tailwind design-system integration or rule findings parity; zero rules become fully helper-ready from this helper alone.

## Consumers

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

## Evidence and scope

The capture is from the real helper invoked by the original six consumer test files, with their original rule assertions intact. Tailwind 4.3.3 is installed in scratch with scripts disabled; a Go overlay replaces only the two developer-specific fixture search-root constants and wraps the helper to record its actual root and fileExists probes. No upstream authored file or shared harness was edited. The independent comparator calls the unmodified Go helper again. Native and Node receive only root/file-presence metadata, never projected rule verdicts. Captured results and traces are retained separately for audit. Control cases exercise all 64 candidate-presence masks over 12 roots, including relative dot segments, empty roots, doubled slashes, Unicode and Linux backslash behavior.

The owned package test replays the recorded consumer metadata and controls without a network dependency. Current Go consumer assertions were run live at capture time; the replay does not claim a new live fixture run. Baseline and mutant exits are zero and stderr is empty. Source, emitted and sanitizer builds execute the same .a implementation.

Reproduce: source /workspace/adamic-tools/env.sh; run go test ./stage1/cohere/lint/helpers/from_wave1_11 -count=1 -v -timeout=10m with output redirected to a log. For fresh live capture, install tailwindcss@4.3.3 under a scratch prefix, then validate.py --scratch <scratch> --package-root <prefix>. Test output always goes to files.

Setup printed Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 84s and done 84s. nproc=5, cpu.max=400000 100000. Full repository gate was not run; scoped owned package comparison and vet passed. Previous rule work remains pushed at 0fe63020; the shared Program/CSS, multiple-fix and compiler/config-base gaps remain blocked and explicit in those reports.
