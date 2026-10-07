Built: six earlier wave-05 ports complete; three newly claimed React hook rules are blocked and unported.
Commits: output ports 948dd987, evidence b9b56e43 and 9d69fd77; new claims ff27ca5e, all pushed.
Commands and outputs: prior output verifier PASS; React gap verifier PASS with Go findings and native JSX refusal for each rule.
Mutants: two output verdicts and CFG edges caught by Go bytes; three raw-fact guards and safe-name regression caught directly; no-op parser probe caught by refusal assertion.
Not covered: the three React ports, their rule mutants/corpus parity/timings, full gate/upstream matrix, or shared emitted-JavaScript harness.

# Continuation 2: exact frontend blocker

After the previous six rules were completed, tested and pushed, all origin heads were fetched. The combined 197-rule ranking, 30 unique origin claim blobs reserving 137 ranked names, and 25 ranked baseline implementations selected the next three:

- `react-hooks/globals`
- `react-hooks/immutability`
- `react-hooks/no-deriving-state-in-effects`

The claim update `ff27ca5e` was pushed before writing the private probe or oracle. Seven high-volume apparent candidates were skipped because their baseline implementations use `Rules.add`/helper reporting; their concrete source locations and every scanned origin tip are in [selection.json](evidence/selection.json). The 26th baseline port, method-signature-style, is outside the 197-rule checker-dependent ranking.

The shared native parser at `stage1/typescript/parser/parser.ts` cannot parse an ordinary self-closing JSX element. The wave-owned [jsx_probe.a](gaps/jsx_probe.a) builds successfully but exits 70 at runtime on:

```typescript
let globalCount=0;
export function Component(){globalCount=1;return <div />;}
```

The observed diagnostic is:

```text
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 72 in /workspace/wave-05-react-probe/input.tsx
```

This is a frontend refusal, before native rule judgments run. An independent Go loader parses that same `.tsx` file and invokes the unchanged production rules; it exits 0 with the globals diagnostic at bytes 46..57, message `globalReassignment`, and zero fixes/suggestions. The retained verifier reproduces the same refusal separately with one positive Go control for each of the three claims: global reassignment, mutation of a props field, and a typed useEffect-derived setter call. Each Go rule reports; native parsing exits 70. The typed fixture uses the exact React declaration stub from the pinned production Go tests.

A non-JSX helper control exits 0 natively with `jsx nodes 0` and is clean under Go, demonstrating that the executable and program loader work. Removing the probe's `parser.file()` call builds and exits 0 with empty stderr; the expected-refusal assertion rejects it. That is a probe mutant, **not** a mutant of an implemented React rule.

Reproduction, with subprocess output written directly to files:

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_05_react/validate_gap.py > /workspace/wave-05-react-gap.log 2>&1
```

The parser is outside this wave's allowed rule directories. Ahra's earlier correction says, "If anything else blocks you, say exactly what it is and stop, rather than editing shared files." This is a shared frontend gap, rather than the shared harness gap covered by the instruction to port the rest and move on. Work stops here under that instruction. No parser, compiler, shared registration generator or shared harness was edited. No React rule is advertised as implemented, and no new rule mutant, corpus equality, release/sanitizer result or timing is claimed for these three. Their reservations remain explicitly blocked.

The completed six ports and their validations are unaffected. The final output ports' corpora, 60 positive/negative programs, sanitizers, release checks, mutants and native/Go measurements are in [OUTPUT_REPORT.md](../wave_05_next/OUTPUT_REPORT.md). Setup remains the previously successful 82s run; nproc is 5.
