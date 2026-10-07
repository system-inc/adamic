Built: tailwind.FindEntryPoint and collapse.ValueToCss, one .a file each; each removes a dependency from the same six Tailwind rules, zero fully helper-ready rules.
Commits: FindEntryPoint claim 48292211 and implementation 2bf8e499; ValueToCss claim 8e5bc04d pushed before implementation; this report accompanies its implementation commit.
Commands and outputs: all six original Go consumer suites plus two targeted finding probes PASS; 922 entry-point cases/130,833 bytes and 233 value-tree cases/13,040 bytes match Go on source Node, emitted JavaScript and sanitized native; owned package test PASS (6.505s), vet PASS.
Mutants: entry_point_priority_reversed and function_closing_parenthesis_omitted compile and run with zero exits/empty stderr, then only byte comparison catches each on all three backends.
Not covered: Windows paths, arbitrary unbounded tree depth, full Tailwind integration/rule findings parity, or the full repository gate; old rule Program/CSS, multiple-fix and compiler/config-base gaps remain explicit.

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

## ValueToCss evidence

Freshly fetched all 417 origin refs and audited 17 distinct claim contents before selecting the unclaimed six-consumer serializer. It preserves word and separator strings verbatim, emits nested function names and parentheses in order, and ignores unknown node kinds with their children, matching the Go switch. The Go oracle invokes the unchanged public ValueToCss. Each of the six original consumer test files is run independently with its original assertions. An owned overlay adds canonical padding and shorthand logical-padding probes using calc(1px+2px), each asserting its actual expected rule finding. These were necessary because the original canonical and shorthand tests did not exercise ValueToCss at all; that absence initially failed the capture coverage assertion.

Observed helper calls by consumer: canonical 3, class order 2, variant order 1, shorthand 1, conflicts 12, unknown classes 59; total 78. These captures carry the actual Go ValueNode trees, not findings or expected serialization. Source Node and compiled runtimes construct their own trees from that input. The additional 155 controls cover empty/nil sequences, empty strings, separator preservation, Unicode/control characters, ignored unknown kinds, nested functions and depths 1 through 64. The replay explicitly checks the Go helper again. The missing-parenthesis mutant is compiled independently and survives execution before the comparator rejects its bytes. Full outputs, original live Go test logs and summaries are retained under evidence/value. There are no authored changes to shared tests, registration, compiler or upstream sources.

Reproduce live serializer capture with validate_value.py --scratch <scratch> --package-root <scratch npm prefix>; hermetic replay is included in the owned Go package test. This is helper parity over the captured cases, not six complete rule ports or arbitrary-depth proof.
