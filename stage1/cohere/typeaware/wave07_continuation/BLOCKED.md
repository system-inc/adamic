Built: the original wave-07 trio remains complete and pushed; no continuation rule is ported.
Commits: original completion 6264bab2; continuation claim 59bdcd34, pushed before implementation.
Commands and outputs: fetched every origin head, inspected 325 remote refs and the existing checker dispatcher.
Mutants: original wave mutants remain recorded in WAVE_07_REPORT.md; no continuation mutant was run.
Not covered: all three continuation implementations and their finding, sanitizer and timing gates; shared dispatcher access blocks the first rule.

# Continuation status

Branch: `codex/typeaware-wave-07`. The original no-undef-init, prefer-for-of and
consistent-indexed-object-style ports are complete in 82eec308 and were pushed,
with their report and evidence in 6264bab2, before selecting more work.

The continuation claim, 59bdcd34e807795fa50933d1989b6a54b021510f, reserves:

1. nexus/correctness-no-process-exit-after-output.
2. nexus/correctness-no-uncleared-race-timeout.
3. nexus/correctness-require-blocking-standard-streams.

Selection fetched all origin heads and inspected 325 remote refs. The combined
197-rule count ranking excludes the 26 base ports, ports on origin/main at
ef3d907ecdc4c771b016f7d9c52372def057a340 and origin/codex/tsgo-c-library at
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6, and exact rule names in Markdown
claims on every origin branch. The claim scan found 96 ranked names. The three
reserved rules are the first eligible names, all with zero corpus findings.
The claim commit was pushed before any implementation code. No more rules are
claimed after discovering this blocker.

## First blocker: resolved call declaration

The production first rule's `followedCallee` uses the exact declaration returned
by `TypeChecker.GetResolvedSignature(call).Declaration()`. It then inspects the
callee's body, source-file module status and function flags. It declines overload
signatures without bodies, generators, unawaited async functions, `never` returns
and global-script functions outside the current file. Those distinctions decide
whether an indirect call counts as a write before an exit.

The current bridge has `signature`/`signature-shape`, `call-returns`, and symbol
and declaration metadata. Its signature branch returns callee/parameter type
facts, not the selected signature's declaration. `node-symbol-details` returns
a symbol's declarations, which cannot substitute for the exact declaration
selected by overload resolution. The existing type-reference-graph projection
also does not expose resolved call declarations or function bodies.

This missing fact is separate from the shared harness's `.a` module support or
suggestion serialization. Changing source extensions would not supply it.
A new raw `call-declaration` question could return selected declaration identity,
source path, byte span, syntax kind and function metadata, leaving every lint
judgment in Adamic. But a new Go implementation file cannot make it callable:
`Program.Inspect` dispatches questions through the switch in the shared
`bridge/tsgo/checker/facts.go`. No external registration hook exists in this
checkout. Integration therefore needs a shared dispatcher registration as well
as the new question and decoder files.

Ahra's correction says, "If anything else blocks you, say exactly what it is and
stop, rather than editing shared files." Work stops at this first blocker. No
shared dispatcher, registration generator or existing test harness was changed
in this continuation. No speculative rule implementation, empty-finding stub or
unsupported syntax workaround was added. The reservations remain marked blocked,
not complete.

## Evidence and limits

The source locations inspected are:

- `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go`,
  `followedCallee`, starting around line 315.
- `bridge/tsgo/checker/facts.go`, the `Program.Inspect` question switch and
  `signature`/`signature-shape` branch around line 636.
- `bridge/tsgo/checker/declaration_facts.go`, the symbol declaration serializer.
- `bridge/tsgo/checker/type_reference_graph.go`, the type-syntax projection.

The fetch command was:

```sh
git fetch origin '+refs/heads/*:refs/remotes/origin/*' > /tmp/wave-07-fetch-next.log 2>&1
```

It exited 0. The remote claim push also exited 0. Selection and capability checks
were read-only Git and source inspections. This is a source-level capability
finding, not a runtime refusal measurement. No continuation compilation, byte
comparison, rule mutant, released-handle check, sanitizer run or native/Go timing
was performed. Zero frozen-corpus counts are selection observations, not evidence
of agreement for these unimplemented rules. The original wave's successful gates
are preserved in ../WAVE_07_REPORT.md and ../validation-wave-07/.
