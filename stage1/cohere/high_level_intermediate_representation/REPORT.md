# Unit 2 report — stopped on native optional-boolean field read

Unit 2 is unfinished. This stopped checkpoint is on `stage1-hir/wip`, which merges
lint area `ad7bd06632f119abc7680719ad3a7d3b71100f58` at `c7dab335` without rebasing.
Static-components remains unregistered: **0 upstream rule cases certified**.

## Parser recovery and file-kind checks

The parser owner's `88f4a83d` recovery is imported through the lint-area merge.
HIR made no parser change. The shortest former blocker is
`function F(){return <x>{a}{b}</x>}`, parsed as **TS**. It now parses on native and
Node; typescript-go and Node print byte-identical whole ASTs and diagnostics.
The retained trees include diagnostic 1005 at 26 (`';' expected.`) and 1110 at 30
(`Type expected.`). See `validation/unit2-merge/shortest-{go,node}.tree`.

The oracle captures Go's `source.ScriptKind`, selected from each case filename.
Its exported files retain `.ts`, `.tsx`, `.js` and `.jsx`; direct and cached replay
pass those paths into the imported parser. The JSX-looking `rangesFor` tests use
`fixture.ts` in Go, so TS recovery is the correct contract for them. Upstream TSX
cases retain TSX. There is no remaining parser gap on this input.

## Coverage and mutants

| Certificate | Original functions | Probes | Overall |
| --- | --- | --- | --- |
| Last native = Node checkpoint, typed/async forms | **966 / 1,465** | **63 / 63** | 1,029 / 1,528 |
| Current Node, structural optional chains | **1,013 / 1,465** | **64 / 64** | 1,077 / 1,529 |
| Current native | compilation blocked | compilation blocked | no current-tip execution certificate |

All original functions remain in the census. The 23 Flow graphs, 40 excluded
Flow upstream fixtures, and 45 classified original Go test skips remain intact.
No admitted failing graph was dropped. Every admitted graph in each passing
certificate compares byte for byte with Go's hir-v1 dump through direct and
cached ForFunction construction.

Casts, satisfies, regex recovery, deletes, typed declarations, typed/default/rest
parameters, async/await and generator metadata now have native/Node certificates.
Optional dot chains preserve shared alternate blocks and inherited continuation
kinds; computed optional loads preserve Go's instruction flag. Method calls keep
Go's receiver/property evaluation and exact optional-marker child roles.

**69 / 69 semantic mutants pass on current Node**, including both optional-path
mutants. The seven new cast/delete/function-form mutants passed on native and
Node before the Optional extension; 67 semantic mutants have native certificates
across checkpoints, but the combined current native matrix cannot run. Earlier
native certificates and their scope remain described in EVIDENCE.md. The checked
FunctionIndex off-by-one mutant stopped on both runtimes at the last native
checkpoint. Brand, private-mint and owner checks remain unchanged.

The three moved mutant anchors were refreshed. Await and typed-binding mutants
now preserve nested graph construction, so they reach a semantic disagreement
rather than aborting with a missing nested path.

Logs are retained under `validation/unit2-merge/`: expressions/native/Node,
function forms/native/Node, whole Node mutant matrix, and native refusals.

## Language gap for @system_adamic

Go `terminal.go:217` defines `Optional`; line 218 is `Optional bool`. The Adamic
tagged terminal record in `core.ts:79` contains `readonly optional?: boolean`.
Native refuses the field read in `dump.ts:108:145`:

```
adamic: /workspace/adamic/stage1/cohere/high_level_intermediate_representation/dump.ts:108:145: stage 0 can't lower a field of type boolean | undefined yet
```

[testdata/optional-boolean-gap.a](testdata/optional-boolean-gap.a) is the checked
three-line reproducer. It prints `true` on Node. Native compilation reports:

```
adamic: /workspace/adamic/stage1/cohere/high_level_intermediate_representation/testdata/optional-boolean-gap.a:2:53: stage 0 can't lower a field of type boolean | undefined yet
```

No representation or field-read workaround was applied. EVIDENCE.md's latest entry
records both this gap and closure of the earlier parser blocker.

## Allocation measurement

The previous shared-index 12-function probe measured **3,212 -> 3,869 allocations
(+657)**, with frees equal to allocations. Retains rose by 2,168, releases by 1,905,
and peak live objects by 24 (175 -> 199); regions stayed zero. The current Optional
tip cannot be measured natively because its build is refused.

Known contributors to that earlier delta are boxed concrete indices, canonical
handle arrays, and per-function SSA adapter closures. The closures bind the owning
HIRFunction because the imported SSA identifier callback does not receive it.
The measurement does not isolate each contributor; the brand remains intact.
The post-merge **last passing native checkpoint (966 originals + 63 probes)** was
reconstructed in an isolated detached checkout at `c7dab335`; its exact source
delta is retained in `validation/unit2-merge/native-checkpoint.patch`. This is a
historical certificate and measurement, not current-tip native certification.

Its counted build matches Go on **1,029 / 1,528** graphs in the checkpoint manifest. The same 12-function
dump is byte-identical to the previous measured dump. Its metrics are:

- Allocations/frees: **4,124 / 4,124**, **+912** from 3,212 and **+255** from 3,869.
- Retains/releases: **15,849 / 14,727**.
- Peak live: **234**; regions: **0**.
- Full admitted census: **2,181,808 allocations and frees**, peak **11,956**.

The additional +255 accompanies the parser recovery merge and typed/function-form
extension, including per-function metadata callbacks and body-child slices. No
isolated share is claimed for those contributors. Count outputs are retained in
`native-12-counts.txt` and `native-corpus-counts.txt`.

## Remaining work

After the language ruling: complete patterns/destructuring and method/function
variants, remaining construction paths, rule-owned ForFunction and compilation-unit
integration, then static-components in its own directory under
`docs/lint-registration.md`. Finish the full construction and upstream rule
comparisons, native/Node mutant matrix, and allocation measurement before calling
unit 2 complete. The watched plan branch has not been pushed.
