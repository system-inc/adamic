Built: released-reservation audit and a compiled .a probe for unavailable program/module facts; no new rule ports.
Commits: claim 19e28bf4 was pushed before code; blocker evidence is committed separately on the same branch.
Commands and outputs: native options control exits 0; both missing-fact queries exit 70; Go production fixtures PASS 0.045s.
Mutants: replacing each missing-fact query with options exits normally and is caught by the required-refusal assertion.
Not covered: the three new ports, rule mutants, byte parity, sanitizer/released-handle checks or native-versus-Go rule timing; stopped at the scope blocker.

## Released reservations and selection

After fetching all origin heads, 325 refs were present. Explicit releases were
treated as available unless another origin claim actively reserved the rule.
Wave 22's continuation now reserves all three earlier released rules:

- @typescript-eslint/no-misused-promises
- @typescript-eslint/no-misused-spread
- nexus/concurrency-no-lost-update

They were skipped for that active claim, not for their old released claims.
The next three available rules, original remaining positions 97 through 99 in
VOLUME_REPORT.md's combined-volume ranking, all with zero recorded findings:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

Each was checked against every fetched origin Markdown claim and native sources
on origin/main and origin/codex/tsgo-c-library. None was claimed or ported on
those bases. Claim 19e28bf4 was pushed before the probe was written. The complete
pre-claim snapshot is validation-wave-21-released/origin-claims.json.xz.
These new reservations remain claimed but unimplemented; no later claim was made.

## Observed blocker

The Go blocking-streams rule requires the actual program's complete source list
and module resolution, not just the manifest roots or declaration origins.
Its correctnessRequireBlockingStandardStreamsBuildIndex calls:

- Program.SourceFiles(), at lines 223 and 235 of its production Go source;
- Program.ResolveModule(importer, specifier), at line 243;
- Program.GetSourceFileForResolvedModule(...), at line 247.

That index decides whether another program file imports the current file and
whether imports transitively reach StandardStreams.ts. Guessing these from
filenames or a filesystem walk would not preserve Go's decisions. The existing
checker query dispatcher exposes neither the complete source list nor module
resolution. The bridge source API search and refusal-routing location are saved
with the probe results.

The standalone wave_21_program_facts_probe.a builds against the existing native
checker archive. It loads a real program and inspects the same SourceFile for
an existing query and two proposed raw-fact queries:

| Question | Exit | stderr |
| --- | ---: | --- |
| options | 0 | empty |
| program-sources | 70 | adamic: panic: unsupported checker question: program-sources |
| module-resolution-graph | 70 | adamic: panic: unsupported checker question: module-resolution-graph |

The live-checker control excludes invalid handles, paths and source positions
as the cause. Both placeholder-query mutants return the options frame with exit
0 and empty stderr; the required unsupported-query assertion catches each.
These are blocker-probe mutants, not mutants of unimplemented rule judgments.

The independent Go production tests for blocking standard streams pass their
real-site, positive and negative cases, including imported scripts and
transitive blocking. This confirms that program/module facts are part of the
reference behavior, rather than hypothetical capabilities sought by this port.

## Scope and stop

Ahra's correction says, "Keep your changes inside your own rule directories"
and "If anything else blocks you, say exactly what it is and stop, rather than
editing shared files." Adding these raw facts would require a new registration
in the shared bridge/tsgo/checker/facts.go dispatcher, outside the rule files.
Under that scope restriction this continuation stops at the missing bridge
facts. No dispatcher, shared registration generator, shared test harness or
protected compiler source was edited in this continuation. A future authorized
bridge extension could supply the facts; this is not a claim of permanent
impossibility or a completed port.

The source reference sizes are 685, 366 and 975 Go lines respectively. No
constant-silent or partial native judgments were substituted for them.

## Reproduction

The previously installed toolchain and archives were reused; setup was not
rerun. The original successful setup took 81 seconds and nproc was 5.

```
source /workspace/adamic-tools/env.sh
/workspace/wave21-next-artifacts/adamic build \
  stage1/cohere/typeaware/wave_21_program_facts_probe.a \
  -o /workspace/wave21-program-facts-probe \
  --tsgo /workspace/wave21-next-artifacts/checker.a \
  > validation-wave-21-released/probe-build.stdout \
  2> validation-wave-21-released/probe-build.stderr

# Run the binary on an ASCII "export {};" SourceFile with its final newline,
# passing options, program-sources and module-resolution-graph respectively.
# Exact argv, exits and streams are represented by the probe source/results.

cd cohere
go test ./internal/lint/rules/nexus \
  -run '^TestCorrectnessRequireBlockingStandardStreams(RealSites|Fires|StaysSilent)$' \
  -count=1 -v > ../stage1/cohere/typeaware/validation-wave-21-released/go-production.log 2>&1
```

The test output was sent to a log, never piped. The mutant uses the same probe
source with `const question = 'options';` in place of its argument-selected
question. All evidence is in validation-wave-21-released. The full gate was
not run because no rule or shared implementation was changed.
