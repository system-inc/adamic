Built: independent three-rule Go oracle and a native .a frontend capability probe; the three React rule ports are blocked.
Commits: previous fifteen ports and evidence pushed through c47e2d1e; new claim 9e54b141 pushed before the probe/oracle work on codex/typeaware-wave-23.
Commands and outputs: native probe builds; non-JSX control exits 0; all three JSX controls exit 70; Go accepts them and emits three findings; Go compiler/repository baselines emit zero findings; vet, formatting and source diff checks pass.
Mutants: no new native rule mutant, sanitizer gate or released-handle test is claimed because no new native rule implementation is complete; previous fifteen ports retain their pushed evidence.
Not covered: native findings/fixes/suggestions parity, React rule decisions, graph lowering, rule mutants or native-versus-Go timing for this batch. The shared JSX parser is the confirmed blocker.

The previous fifteen claims are fully implemented, tested and pushed. An initial
push reported Everything up-to-date. The all-heads fetch succeeded, producing
389 origin refs. Main advanced to e011f8f60899586d6373a5ccb07335ad82cfbf3c without
changing its typeaware port inventory. The bridge tip remains
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Selection checked the same 26 production
ports and 142 claimed ranked rules across 33 distinct Markdown claim blobs.
There were 30 remaining ranked rules. The first three are:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

All three have zero counts in the original compiler and repository ranking.
Earlier candidates, including prefer-numeric-literals, prefer-object-has-own and
prefer-object-spread, are claimed elsewhere and were skipped. Neither full nor
short names of these three appear in any fetched claim. The claim push completed
before implementation work. [selection.json](validation-wave23-react-blocker/selection.json)
preserves refs, claim texts, exclusions and the remaining ranking. No more rules
were claimed. These three claims are explicitly marked blocked and remain reserved.

The concrete blocker is JSX grammar in stage1/typescript/parser, rather than .a
module loading in the shared lint harness. The scanner has scanJsx and JSX token
support, but the native parser has no JSX productions. The shared harness branch
codex/lint-harness-dot-a also has no JSX grammar in its parser.ts. A native probe
built from [wave_23_react_parser_probe.a](wave_23_react_parser_probe.a) loads the
same checker program and runs the unchanged native parser over a source file.
The ordinary TypeScript control returns parsed SourceFile with exit 0. Each
parse-valid JSX control fails before any rule can run:

| Rule control | Native exit | Native error | Independent Go |
| --- | ---: | --- | --- |
| static component creation and Inner tag | 70 | expected GreaterThanToken, got SlashToken at 53 | staticComponents at bytes 47..52 |
| synchronous effect setter, followed by div tag | 70 | expected GreaterThanToken, got SlashToken at 101 | setStateInEffect at bytes 74..82 |
| unconditional render setter, followed by div tag | 70 | expected GreaterThanToken, got SlashToken at 83 | setStateInRender at bytes 59..67 |

The Go oracle in [oracle_wave_23_react.go](testdata/oracle_wave_23_react.go) loads
its own TypeScript program and runs all three unmodified production rules. It
imports no bridge code. The same tsconfig and source files are used by the native
probe. All source files have zero Go parse diagnostics. Minimal declaration roots
provide the production rule's Dispatch alias and hook signatures for the setter
controls. The Go stream has three complete findings, zero fixes and zero
suggestions. Its compiler baseline covers all 77 frozen roots and its repository
baseline all 287 frozen roots; both have zero findings. These Go-only observations
are not native parity or completion of any port.

The rule implementations additionally consume cohere's lowered function arena:
SSA values and phi nodes, capture-namespace translation, reverse-postorder blocks,
post-dominance for unconditional calls, control-dependence for ref guards, and
manual-memoization erasure for effects. Those passes have not been translated in
this batch. No syntax-only substitute was implemented or represented as a port.
The confirmed JSX failure already prevents the required complete native frontend
path and positive parity checks, so execution stops before changing shared parser
files. This follows the instruction to stop on blockers outside the worker's
territory rather than editing shared files. The .a-loader/suggestion-serialization
work alone does not close this parser gap.

Exact commands, run from the repository root after sourcing the existing toolchain:

```sh
/workspace/wave-23/behavior-final/adamic build stage1/cohere/typeaware/wave_23_react_parser_probe.a -o /workspace/wave-23/react-blocker/parser-probe --tsgo /workspace/wave-23/behavior-final/checker.a > /workspace/wave-23/react-blocker/native-build.log 2>&1
/workspace/wave-23/react-blocker/parser-probe /workspace/wave-23/react-blocker/tsconfig.json /workspace/wave-23/react-blocker/control.a > /workspace/wave-23/react-blocker/native-control.stdout 2> /workspace/wave-23/react-blocker/native-control.stderr
# Repeat the native command for component.tsx, effect.tsx and render.tsx.
# Each exits 70 with its recorded parser error.
go -C cohere build -overlay /workspace/wave-23/react-blocker/oracle-overlay.json -o /workspace/wave-23/react-blocker/go-oracle /workspace/adamic/cohere/adamic_wave23_react_oracle.go > /workspace/wave-23/react-blocker/go-build.log 2>&1
/workspace/wave-23/react-blocker/go-oracle /workspace/wave-23/react-blocker/tsconfig.json /workspace/wave-23/react-blocker/all.manifest > /workspace/wave-23/react-blocker/go-all.stdout 2> /workspace/wave-23/react-blocker/go-all.stderr
/workspace/wave-23/react-blocker/go-oracle /workspace/wave-23/typescript/src/compiler/tsconfig.json /workspace/wave-23/compiler.manifest > /workspace/wave-23/react-blocker/go-compiler.stdout 2> /workspace/wave-23/react-blocker/go-compiler.stderr
/workspace/wave-23/react-blocker/go-oracle /workspace/adamic/tsconfig.json /workspace/wave-23/repository.manifest > /workspace/wave-23/react-blocker/go-repository.stdout 2> /workspace/wave-23/react-blocker/go-repository.stderr
go vet ./... > /workspace/wave-23/react-blocker/vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/react-blocker/gofmt.log 2>&1
git diff --check > /workspace/wave-23/react-blocker/diff.log 2>&1
```

An initial probe command used the cohere submodule working directory and therefore
failed to find the relative native source. The corrected repository-root command
built successfully; only its successful-build capability results above are used.
No build failure is counted as a mutant kill. This report adds no new performance
claim: timing a failing native parser against successful Go lint is not comparable.
No full test gate or new sanitizer run is claimed for an unimplemented rule batch.

The existing setup succeeded earlier in this session: Go 1.27.1 ready 0s;
clang 20.1.8 ready 0s; Node 24.19.0 ready 0s; submodules 0s; cache warm 83s;
setup done 83s. nproc was rechecked and reports 5. The original log is retained.
Cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db and typescript-go remains
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No pins changed.

[validation-wave23-react-blocker](validation-wave23-react-blocker) contains the
complete compressed Go streams, native errors and output, source contents/hashes,
configuration, exits, selection snapshot and logs. The new native source is .a;
.tsx names identify TypeScript input controls stored as data, not Adamic ports.
No shared parser, registration generator, test harness or protected compiler file
changed. No PR was opened.
