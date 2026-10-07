Built: rebased twelve claimed rule ports onto area/stage1-lint; integrated JSX now passes positive controls.
Commits: prior pushed tip 717b356c; integration base 7481e032; tested source commit is pinned in evidence metadata.
Checks: all 17 landing steps, 131 JSX controls, both corpora, sanitizers and released-handle guards pass.
Mutants: twelve rules, three JSX dispatches, nine legacy listeners, graph/fact/data/ownership guards and shared harness witnesses are caught.
Not covered: three parked React analysis ports, four production bridge routes, shared type-aware context integration, own emitted-JavaScript comparison and the full repository gate.

The latest rebase and required-input checks are in [REQUIRED_INPUTS_REPORT.md](REQUIRED_INPUTS_REPORT.md).

## Integration outcome

The branch was rebased cleanly onto origin/area/stage1-lint at
7481e0324e34a2537aafa9db7eeacda50405611b, containing origin/main
39638d9e278d38bb5aeae887f46d55a70e47aaad. Incoming shared harness,
finding-model, JSX parser and allocator helper changes were retained. No shared
files, protected compiler files or rule bodies were changed by this refresh.
Only codex/typeaware-wave-07 is pushed, never main or an area branch.

The owned JSX validation now uses actual integrated parser sources rather than
an isolated historical copy. parser-pins.json checks every source byte against
the integration base. The three hook parser controls now exit 0 with jsx 1;
the ordinary control prints jsx 0. Independent Go still reports all three hooks.
The hook claims remain parked on native high-level IR, SSA/value flow and
capture analysis pending #dnv6f2c; parser support is no longer their blocker.

## Agreement and checks

The original trio passes 283 parse-valid controls, 173 findings and 47,635
identical bytes. Two upstream empty-index-signature recovery cases remain
excluded explicitly by the established gate. Repository: 287 files, 16 findings,
22,753 bytes. Compiler: 77 files, 24 findings, 9,282 bytes. Ordinary and sanitizer
streams agree, including findings, fixes and suggestions.

The next six rules all pass their positive controls, options, both frozen
corpora and sanitizer streams. The regex comparison includes 30 suggestions.
The three JSX rules pass 131 controls: default 65 findings / 28,512 bytes;
element mode 48 / 24,534; globals 63 / 27,829; both 46 / 23,851. All four also
match under ASan/UBSan/LSan. Repository is 18,485 bytes and compiler 5,318 bytes,
with zero JSX findings. These production rules emit no fixes or suggestions.

Compiled JSX declarations, rule.json kinds and production Go Run keys match on
184 bytes. The older nine declared listener maps match on 358 bytes, including
sanitizers. Existing symbol metadata succeeds live and refuses released handles
with panic 70. All 114 pinned library sources match 3,793,522 external bytes
under sanitizers. Source lint passes 276 rules over 15 owned .a files.

The owned landing runner records 17 successful steps: original trio; timer;
process; streams; rest; promise; regex; continuation questions; regex question;
bridge packages; fact decoder/refusals/type flags; Node oracle; shared finding
model; vet; gofmt; integrated React parser; historical published parser control.
The exact command arguments and logs are archived in evidence/landing.
The Node filter passes 31 native-versus-Node fixtures, the one-byte oracle mutant
and 12 iterator-copy refusals. The new inherited_static_field_read is included.
Shared tests exercise emitted JavaScript mismatch detection, complete suggestion
serialization, suggestions beside automatic fixes, .a modules, deterministic
regeneration and descriptor rejection. Their successful gate includes expected
nested mutant failures; these are oracle kills, not unresolved test failures.

## Every mutant and its detector

Each rule mutant compiles, exits 0 and has empty stderr. Only the independent
production Go finding/fix/suggestion bytes catch the following changes:

| Rule | Mutation |
| --- | --- |
| no-undef-init | Shift automatic removal start by one byte |
| @typescript-eslint/prefer-for-of | Invert accepted-body predicate |
| @typescript-eslint/consistent-indexed-object-style | Change message punctuation |
| nexus/correctness-no-uncleared-race-timeout | Change unclearedRaceTimeout ID |
| nexus/correctness-no-process-exit-after-output | Change exitAfterOutput ID |
| nexus/correctness-require-blocking-standard-streams | Change exitBeforeBlockingStandardStreams ID |
| prefer-rest-params | Change preferRestParams ID |
| prefer-promise-reject-errors | Change rejectAnError ID |
| prefer-regex-literals | Change unexpectedRegExp ID |
| react/jsx-fragments | Change preferFragment ID |
| react/jsx-no-undef | Change jsxIdentifierNotDefined ID |
| react/no-adjacent-inline-elements | Change inlineElement ID |

Additional valid exit-0 mutants and their detectors:

- Each JSX rule's first named listener becomes Unknown. Full Go finding bytes catch missed dispatch in all three cases.
- Nine legacy declarations increment their first key: indexed, forOf, processExit, timeout, streams, undefinedInit, promise, regex and rest. Production Go listener bytes catch each.
- Native reference-graph symbol equality is inverted, and the Go graph symbol becomes zero. Full production findings catch both.
- symbol-value zeros resolved symbol identity; flow-edge drops successors; module-target erases target filenames; character-span drops regex characters. Full production finding/suggestion bytes catch each.
- Released-program registry deletion is removed in three independent gates: original graph, continuation questions and regex question. Required panic 70 catches each valid exit-0 mutant. Every relevant question separately succeeds live and refuses after release.
- A library copyright character changes. The independent pinned 3,793,522-byte library oracle catches it.
- The filtered core one-byte mutant is caught by Node disagreement.
- Shared emitted-JavaScript mismatch is caught by ordinary Go comparison; the shared second suggestion edit range mutant is caught on Node, emitted JavaScript and native. Registry malformed-descriptor witnesses and ambiguous .ts/.a entries are rejected as specified in their log.

Disk/setup failures are not counted as mutant kills.

## Native versus Go time

Three quiet separate-process pairs per JSX rule after heavy checks finished.
Every timed output is also compared in full. Times include startup, checker
creation, parsing and handling; they do not isolate handler cost.

| Corpus | Rule alias | Native seconds | Go seconds | Native/Go |
| --- | --- | ---: | ---: | ---: |
| repository | fragments | 0.328934 | 0.135969 | 2.42 |
| repository | undef | 0.323819 | 0.135205 | 2.40 |
| repository | adjacent | 0.324053 | 0.136873 | 2.37 |
| compiler | fragments | 2.156862 | 0.325821 | 6.62 |
| compiler | undef | 2.245242 | 0.322136 | 6.97 |
| compiler | adjacent | 2.059332 | 0.319350 | 6.45 |

Earlier ports' fresh sequential gate observations (one process pair each):

| Unit | Repository native / Go seconds | Compiler native / Go seconds |
| --- | ---: | ---: |
| Original trio together | 0.330205 / 0.139940 | 2.105856 / 0.302114 |
| timer | 0.306065 / 0.143409 | 1.762593 / 0.377163 |
| process | 0.316711 / 0.159969 | 1.883600 / 0.317194 |
| streams-final | 0.375847 / 0.149040 | 1.815013 / 0.308390 |
| next-rest | 0.303144 / 0.141065 | 2.292603 / 0.358574 |
| next-promise | 0.288831 / 0.139348 | 1.772812 / 0.302425 |
| next-regex | 0.282488 / 0.147963 | 2.078781 / 0.319543 |

Native remains slower. No speedup is claimed.

## Setup and recovery

The first setup failed in cache warming with mkdir .generated: no space left
on device. Go's disposable build cache occupied 28 GB of the 32 GB filesystem.
After authorized go clean -cache, setup passed: Go ready 0s, clang ready 1s,
Node ready 1s, submodules ready 1s, cache warm 264s, done 264s. nproc is 5,
cgroup cpu.max is 400000 100000, memory 17.6 GB. Source
/workspace/adamic-tools/env.sh before each command.
Repository files and saved evidence were preserved. The runner now deletes only
completed owned ELF/archive scratch outputs, retaining compiler/archive inputs
needed by downstream guards. Recovered sizes and paths are logged.

## Exact remaining gaps

The shared RuleContext at the integrated revision still has no checker-program
handle. These standalone type-aware handlers are not registered with that
context. Four raw bridge dispatcher routes remain private overlays:
wave07-symbol-context, wave07-control-flow, wave07-program-modules and
wave07-regex-structure. Shared-file ownership prohibits adding those dispatcher
cases here. Both private behavior and unregistered refusal are tested. The
shared .a/profile/suggestion harness itself has landed and its focused tests
pass; its absence is no longer a blocker. Own type-aware emitted-JavaScript
comparison and the full repository gate were not run. Three analysis-dependent
React hook claims remain parked. No new claims were made in this refresh.

Historical parser/harness blocker statements in earlier reports describe their
then-current bases; this report supersedes their current-status claims.
