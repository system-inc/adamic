# Performance's form-scoped host-any row (#m36zjrk)

`TestHostAnyModelledPerformanceForms` reads Modelled rows through
`hostAnyModelledRows`. There must be one performance row with a Forms column.
The named `performance.<name>` reads/calls must match exactly the registered
`node:perf_hooks.Performance.*` members. The two typeof lines remain forms,
without adding a seventh member. The namespace performance object is registered,
no perf_hooks.d.ts location enters `hostAnyModelled`'s declaration-wide allow-list,
and every repository evidence path exists.

| Mutation | Named pin | Catch line |
| --- | --- | --- |
| Add `toJSON` to library_node_process.go's loop registering the six Performance members | `TestHostAnyModelledPerformanceForms` | `performance Forms do not match registered members: documented=[clearMarks clearMeasures mark measure now timeOrigin] registered=[clearMarks clearMeasures mark measure now timeOrigin toJSON]` |
| Same seventh-member registration | `TestHostAnyAudit` | `node:perf_hooks.Performance.toJSON: value-carrying any at node:perf_hooks.Performance.toJSON.signature[0:0].return` at `perf_hooks.d.ts:116` |

`TestHostAnyPerformanceToJSONMutantIsCaught` uses `buildcache.GoTestMutated` to
build the private mutant, then runs each named pin separately from that binary.
Each must exit 1 with its catch line; the audit must also name perf_hooks.d.ts:116.
A build error or a missing fixture cannot substitute for either catch.
`TestProduct_HostAnyPerformanceToJSONMutant` prepares the same product.

This pin is one commit above the separate performance-table docs commit
`9b68c851`, on Runtime's `runtime/performance-row-5` (`3008db8c`). Its base already
contains performance_provided/performanceCore.ts. Validation runs directly on
this stack; no scratch merge or copied Runtime fixture is needed.
