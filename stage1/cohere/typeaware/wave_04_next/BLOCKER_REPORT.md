Current status: this historical blocker report is superseded by [REPORT.md](REPORT.md). The isolated adapter now runs all three continuation ports; normal and sanitized byte comparisons pass. Earlier failed push commits were subsequently pushed.

Built: the original three rules remain complete; continuation has two prepared raw checker questions and `.a` decoders, not rule ports.
Commits: original completed tip `67b15e35`; continuation claim `940df421`; blocker evidence committed separately.
Commands and outputs: checker tests PASS; both Adamic probes compile; native metadata control exits 0 with 111 bytes.
Mutants: original three rule mutants remain validated; no continuation rule mutant or parity result is claimed.
Not covered: all three continuation rule implementations, corpus comparisons, sanitizers and timings are blocked on bridge dispatch authorization.

# Continuation status

The first three unclaimed rules after fetching all 325 origin references are:

- `nexus/correctness-no-process-exit-after-output`
- `nexus/correctness-no-uncleared-race-timeout`
- `nexus/correctness-require-blocking-standard-streams`

All have zero findings in the existing frozen volume inventories. Zero volume is
selection evidence, not evidence that a native implementation works. These rules
remain claimed and unfinished. No additional rules were claimed.

## Observed blocker

Ahra's correction says: "Keep your changes inside your own rule directories" and
"If anything else blocks you, say exactly what it is and stop, rather than editing
shared files." The latest request permits isolated new bridge questions but does
not explicitly permit changing shared dispatch. The original wave allowed single
registration lines; clarification was requested before making any new such edit.

Existing `node-symbol-details` and `declaration-details` expose one parent, but
omit the ancestor flags/global augmentation and source external-module bit these
rules use to distinguish true platform globals and `NodeJS.Process` members.
Existing signature questions expose types and parameters but omit the selected
signature declaration and body. No existing dispatch case supplies those facts.

Two new Go files now expose raw compiler facts, with matching Adamic decoders in
this directory. Direct Go tests execute both functions against a positive `.a`
control. They verify the external-module bit, declaration identity, void return
flags and selected function body. Existing checker regression tests also pass.
Neither question makes a lint verdict or produces an edit.

The native `bridge_probe.a` compiles and runs the existing `declaration-details`
question: exit 0, 111 stdout bytes, empty stderr. Each new request exits 70,
produces no stdout, and prints `adamic: panic: unsupported checker question:` with
its name. Thus the compiler capability exists, but it is not reachable through
the registered public bridge. The matching decoder build probe compiles with
empty build stderr; decoder execution through the new questions remains blocked.

The minimum pending integration change in `bridge/tsgo/checker/facts.go` is:

```go
case "wave04-next-declaration-context": return p.wave04NextDeclarationContext(out, c, node, question)
case "wave04-next-resolved-signature-declaration": return p.wave04NextResolvedSignatureDeclaration(out, c, node, question)
```

These lines were not applied, even through a build overlay. The shared registration
generator and existing test harness were not edited. This is a shared-file scope
blocker, not an automatic approval rejection or a `.a` module loading failure.
The blocking-streams rule additionally needs import type-only/computed-loader
metadata absent from the existing `program-imports` response; its bridge extension
and native CFG work have not been started after hitting this blocker.

## Reproduction

```sh
source /workspace/adamic-tools/env.sh
go test ./bridge/tsgo/checker -count=1 -v > /workspace/typeaware-wave-04-next/checker-final.log 2>&1
/workspace/typeaware-wave-04/final/adamic build stage1/cohere/typeaware/wave_04_next/bridge_probe.a -o /workspace/typeaware-wave-04-next/bridge-probe --tsgo /workspace/typeaware-wave-04/final/checker.a > /workspace/typeaware-wave-04-next/build.log 2>&1
/workspace/typeaware-wave-04/final/adamic build stage1/cohere/typeaware/wave_04_next/decoder_build_probe.a -o /workspace/typeaware-wave-04-next/decoder-probe --tsgo /workspace/typeaware-wave-04/final/checker.a > /workspace/typeaware-wave-04-next/decoder-build.log 2>&1
```

`probe-results.json` records the three native request runs. Probe input is the
exact two-byte source `x;`; config is the existing type-aware testdata config.
The earlier setup succeeded in 77s, with `nproc=5`; its timing log is already
preserved with the original wave report. No new setup or performance claims.

## Delivery

The claim commit `940df421` was pushed successfully before code. The prepared
facts and blocker evidence are committed as `39cdb8de`. Both attempts to push
that commit failed with `fatal: could not read Username for 'https://github.com':
No such device or address`. Read-only remote verification then failed with the
same credential error. This is a missing Git credential, not an automatic
approval rejection. No token was printed or searched for.

A format-patch fallback containing the unpushed continuation commits is saved at
`/workspace/typeaware-wave-04-next/continuation-blocker.patch`.
