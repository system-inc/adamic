# Fourth batch blocked by native JSX grammar

The previous nine default-option ports are tested and pushed at
24d4f417aa8800a6fc22f70fc8eaa8263f5d9a60. After fetching all origin
heads without submodule recursion, the selection audit covered 389 origin refs
and the 197-rule combined by-volume ranking. Excluding ports on main/base and
Markdown claims on every origin ref, the first available rules were:

- react-hooks/set-state-in-effect
- react-hooks/set-state-in-render
- react-hooks/static-components

Their claim update was committed and pushed in e22fc9c1 before implementation.
All three remain pending. This batch publishes a blocker probe and evidence,
not rule implementations or a claim of rule agreement.

## Observed blocker

Every production static-components finding requires a JSX tag. The unchanged
Go rule's compilation gate requires JSX in the unit's body. Native Parser on
this branch parses `<Component />` as a type assertion and refuses it before
rule traversal. Supporting `.a` imports or suggestion serialization in the
shared harness does not implement that missing grammar.

The authored control is gaps/static_component_control.a. A scratch control.tsx
symlink lets the independent Go loader read its JSX, avoiding the already known
Go `.a`-as-TypeScript loading limitation. The native probe reads the same path
and same bytes. Its own executable is compiled from gaps/parser_probe.a and
imports the unchanged native Parser. The independent oracle invokes all three
unchanged production rules with their own loader, walk and file cache, and
imports no bridge implementation.

Observed Go: exit 0, one staticComponents finding at UTF-8 bytes 74 through 83,
zero fixes and suggestions, naming creation at `createComponent`.
Observed native parser: exit 70, empty stdout, with exactly:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 84 in /workspace/wave-27-scratch/fourth-blocker/control.tsx
```

The rule cannot reach a positive comparison or per-rule mutant on this parser.
Ahra's instruction says, "If anything else blocks you, say exactly what it is
and stop, rather than editing shared files." This is a shared-parser blocker,
not the shared-harness module-loader exception. Work stopped here; the two
state-update ports were not started and no further rules were claimed.
No parser, registration generator, harness, bridge or protected compiler source
was changed. There is no silent always-empty static-components substitute.

## Commands and limits

After sourcing /workspace/adamic-tools/env.sh:

```bash
/workspace/wave-27-scratch/third-final/adamic build \
  stage1/cohere/typeaware/wave_27_fourth/gaps/parser_probe.a \
  -o /workspace/wave-27-scratch/fourth-blocker/parser-probe \
  > /workspace/wave-27-fourth-parser-build.log 2>&1
# PASS, then parser-probe control.tsx: exit 70, as above.

# From cohere, overlay virtual main onto this directory's oracle.go.txt:
go build -overlay /workspace/wave-27-scratch/fourth-blocker/overlay.json \
  -o /workspace/wave-27-scratch/fourth-blocker/oracle \
  /workspace/adamic/cohere/adamic_wave27_fourth_oracle.go \
  > /workspace/wave-27-fourth-oracle-build.log 2>&1
# PASS, then oracle tsconfig.json controls.manifest: exit 0, findings 1.
```

Subprocess output and exit records are retained in evidence, alongside the
selection audit, fetch and claim-push logs, config and manifest. The symlink is
scratch only; no authored .ts or .tsx file is committed. The complete authored
control and native probe are committed as .a in gaps.

No new-rule byte agreement, mutants, released-handle tests, sanitizer checks,
corpus gate or native-versus-Go timings were completed in this batch. Previous
nine-rule validation remains in the earlier wave reports. The same wave's setup
passed: Go/clang/Node/submodules ready 0s, build cache warm and done 116s;
nproc 5. This stop does not assert that a future parser or integrated harness
cannot handle JSX; it records the refusal of the current branch's parser.
