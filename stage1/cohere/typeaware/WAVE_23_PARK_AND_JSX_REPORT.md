Built: marked the three HIR/SSA/capture-dependent React claims parked; reserved the next three AST/checker rules and added independent Go controls plus numeric declarations.
Commits: prior landing-ready 569f8d48; parking/new claim da97c734 pushed before new oracle or declaration work; this evidence is committed separately.
Commands and outputs: fetched 529 refs, scanned 33 Markdown claim blobs; Go emits three positive findings; 107 listener bytes match Go/native/sanitized native; vet and diff checks pass.
Mutants: three listener metadata mutants compile, exit 0 with empty stderr and differ from Go at bytes 20, 72 and 99; no new rule-decision mutants claimed.
Not covered: native decisions, finding/fix/suggestion parity and timings for the new batch; shared JSX parsing still blocks execution before any rule can run.

The user's new instruction parks react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components on native
high-level IR, single-assignment and capture analysis being ported on #dnv6f2c.
Their prior oracles, probes, numeric declarations and rule.json manifests were
already pushed. They remain reserved, and count as finished for the landing-first
cap under that explicit instruction. Their native decision ports are not complete.

Current origin/main f8013f0baac41ddc340d76f83bddde38536a8f07 is unchanged and
already included in this worker branch. The prior six gates and expanded Node
oracle passed on that baseline; the remote 569f8d48 was landing-ready. No main or
area branch was written. All origin heads were then fetched. The ranking
combines compiler and repository volumes with lexical ties, excludes production
ports and all fetched reservations. It has 154 claimed ranked rules and eighteen
remaining candidates. The first three are react/jsx-fragments,
react/jsx-no-constructed-context-values and react/jsx-no-undef. Their analyses
use AST and checker queries, without the native HIR/SSA/capture passes.
The claim was committed and pushed at da97c734 before any new code.
Selection is preserved in validation-wave23-react-blocker/park-selection.json.

Each rule now has its own .a numeric listener declaration and a name/kinds
rule.json under listeners-wave23/react. Independent production Go Run methods
produce the exact numeric registrations:

```
react/jsx-fragments	285,286,289
react/jsx-no-constructed-context-values	286,287
react/jsx-no-undef	286,287

```

The independent Go oracle oracle_wave_23_jsx_next.go imports no bridge code and
runs unmodified production rules with its own checker/program. Three parse-valid
positive inputs produce one finding each, with complete serialized messages,
byte ranges, zero fixes and zero suggestions:

```
file	/workspace/wave-23/jsx-next/fragment.tsx
60	78	react/jsx-fragments	preferFragment	Prefer fragment shorthand over React.Fragment. The shorthand `<>` says the same thing with less to read, and a named fragment that carries no props is only the long spelling of it.	0	0
file	/workspace/wave-23/jsx-next/context.tsx
68	70	react/jsx-no-constructed-context-values	defaultMsg	The object passed as the value prop to the Context provider (at line 1) changes every render, so every consumer of this context re-renders on every render of this component even when nothing it reads has changed. Wrap it in a useMemo hook so the identity is stable between renders.	0	0
file	/workspace/wave-23/jsx-next/undef.tsx
26	33	react/jsx-no-undef	jsxIdentifierNotDefined	This JSX tag names a component that has no declaration in scope, so React receives `undefined` where a component was meant and throws at render rather than at build. Almost always a misspelling or a missing import. Import the component, or fix the spelling to match the binding you meant.	0	0
findings 3

```

The declaration-only native probe and ASan/UBSan/LeakSanitizer build both exit 0
with empty stderr and match Go's 107 registration bytes exactly. Each separate
metadata mutant changes that rule's first kind to 0. It compiles and exits 0,
with empty stderr; only Go registration bytes catch it. These are metadata
mutants, not evidence of native rule decisions or React finding parity.
The new .a sources contain no string syntax-kind comparisons or node refetches.
No runtime callbacks are implemented yet and no every-rule/every-node dispatch
is added.

Shared JSX integration remains a concrete frontend blocker. Replaying the
existing current-main native parser probe on these new Go-accepted inputs gives:

```
{
  "native": {
    "exit": 0,
    "bytes": 107
  },
  "native-asan": {
    "exit": 0,
    "bytes": 107
  },
  "react/jsx-fragments metadata mutant": {
    "exit": 0,
    "first_difference": 20
  },
  "react/jsx-no-constructed-context-values metadata mutant": {
    "exit": 0,
    "first_difference": 72
  },
  "react/jsx-no-undef metadata mutant": {
    "exit": 0,
    "first_difference": 99
  },
  "fragment JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 76 in /workspace/wave-23/jsx-next/fragment.tsx\n"
  },
  "context JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got Identifier at 61 in /workspace/wave-23/jsx-next/context.tsx\n"
  },
  "undef JSX probe": {
    "exit": 70,
    "stderr": "adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 34 in /workspace/wave-23/jsx-next/undef.tsx\n"
  }
}

```

The failures happen before lint listeners can run. The current shared parser
also exposes string kinds, and the numeric driver is not present here. Shared
parser/driver/harness files are outside this unit's edit scope. JSX support is
landing on area/stage1-lint; no local shared-file replacement was attempted.
The three new rule decisions remain unimplemented. This report does not count
those three as finished ports or claim native parity. No additional rules were
reserved after this batch and no PR was opened.

An initial declaration build lacked clang in PATH because its shell had not
sourced the toolchain environment. After source /workspace/adamic-tools/env.sh,
both ordinary and sanitized builds passed; that setup failure is not counted
as a mutant kill. Existing setup timings remain Go 0s, clang 0s, Node 0s,
submodules 0s, warm 83s, done 83s; nproc 5. No setup rerun is claimed.

Go build used a worker-owned overlay to compile the new independent oracle
inside the cohere module. It ran on /workspace/wave-23/jsx-next/manifest with
/workspace/wave-23/react-blocker/tsconfig.json, both normally and with
--listener-kinds. Native declaration builds used the current-main compiler
/workspace/wave-23/landing-f801/behavior/adamic; the second passed --sanitize.
All comparison output was saved to files; no test was piped. go vet ./... and
git diff --check exited 0. The full repository gate, full new rule corpus,
new checker released-handle gate, emitted-JavaScript check and native/Go lint
performance comparison were not run for this blocked declaration-only batch.

[validation-wave23-jsx-next](validation-wave23-jsx-next) preserves raw compressed
Go findings, registrations, native outputs, mutations, JSX input controls,
error logs and hashes. No shared finding-model landing sha was supplied, so
no adoption of that unnamed change is claimed.
