# Stage 3 meter 20261123T132637Z

The area label repeats the pinned main tree for reporter compatibility.

| Run | Main SHA | Whole program | Own file | Lowering | tsc entry diagnostics |
| --- | --- | ---: | ---: | ---: | ---: |
| Current | fb672853cf6aeddffe2e2de7d4a3ae95b0df3215 | 79/79 | 79/79 | 2/79 | 0 |
| 20261122T131040Z.tNkys1 | 178b45d36c8989e22f9e8ef4fb611adb6eded24c | 79/79 | 79/79 | 2/79 | 0 |

Lowering, main: 2/79; root-order artifacts: 3.

| Root | Stop | Reason | Node result | Classification | Program result from entry census |
| --- | --- | --- | --- | --- | --- |
| src/compiler/debug.ts | src/compiler/semver.ts:71:9 | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | Node throws here as well; exit 1; TypeError: Cannot read properties of undefined (reading &#x27;assert&#x27;) | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |
| src/compiler/diagnosticInformationMap.generated.ts | src/compiler/commandLineParser.ts:293:35 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |
| src/compiler/types.ts | src/compiler/diagnosticInformationMap.generated.ts:13:45 | reading an enum before its runtime initialization; move the call after the enum declaration | Node throws here as well; exit 1; ReferenceError: __filename is not defined in ES module scope | root-order artifact | Refused src/compiler/binder.ts:496:42: a cast the runtime can&#x27;t check: a cast from { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } to FlowNode, whose member FlowArrayMutation can&#x27;t be seen as { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; }, whose field node FlowArrayMutation keeps readonly and { flags: FlowFlags; id: number; node: unknown; antecedent: FlowNode[] &#124; FlowNode &#124; undefined; } can write (read from the entry census; source order) |

For artifact roots, lowering requires attempted units in every program source, no entry-census finding in the program, and the whole-program passes of every unit measured. No native output is claimed.

Area reuses main's measurement, not a second measurement. main/ was copied to area/ before reporting and gzip; both labels use main-adapted. See compiler-mode.json measurement_sources.

Elapsed: 87m 8s.
