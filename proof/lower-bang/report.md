Head: `a026a05300d81da96b5d5d93e814f5e211995d89`. Main: `37a549e53ea10621ed79ca2387ab2eb6b3109440`.

Loaded `/workspace/adamic-tools/env.sh`, fetched `origin runtime-performance/setblocking-receiver-6`, verified detached head, initialized submodules recursively, and built both revisions with `go build -o /tmp/adamic-{head,main} ./cmd/adamic`. Created `/tmp/adamic-gate` outside the repository.

Sources were saved as `/tmp/p/<name>.ts`. Each revision used `adamic build` (release), `adamic build --sanitize`, and `adamic js`. Generated JavaScript was saved as `.mjs` and run through `node --disable-warning=ExperimentalWarning oracle/node.mjs`, as in the oracle. Source Node runs used that same command with the original `.ts`. Native runs used `ASAN_OPTIONS=detect_leaks=0`. Node version: `v24.19.0`.

Runtime cells show `(stdout, stderr, exit)` with JSON string escapes. “Stopped Dn” means compilation/lowering exited 1 with stdout `""` and the exact stderr in diagnostic Dn; no binary or JavaScript was run. “Lowers” means all three compiler commands exited 0 with empty stderr; native compilation also had empty stdout, while `adamic js` emitted the generated program.

| Revision | Program | Lowering | Node source (stdout, stderr, exit) | Release runtime (stdout, stderr, exit) | Sanitized runtime (stdout, stderr, exit) | JavaScript runtime (stdout, stderr, exit) |
|---|---|---|---|---|---|---|
| head | `class_method.ts` | Lowers | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` |
| head | `interface_view.ts` | Lowers | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` |
| head | `parenthesized.ts` | Lowers | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` | `("3\n", "", 0)` |
| head | `double.ts` | Stopped D4 | `("3\n", "", 0)` | Stopped D4 | Stopped D4 | Stopped D4 |
| head | `subclass.ts` | Lowers | `("derived\n", "", 0)` | `("derived\n", "", 0)` | `("derived\n", "", 0)` | `("derived\n", "", 0)` |
| main | `class_method.ts` | Stopped D1 | `("3\n", "", 0)` | Stopped D1 | Stopped D1 | Stopped D1 |
| main | `interface_view.ts` | Stopped D2 | `("3\n", "", 0)` | Stopped D2 | Stopped D2 | Stopped D2 |
| main | `parenthesized.ts` | Stopped D3 | `("3\n", "", 0)` | Stopped D3 | Stopped D3 | Stopped D3 |
| main | `double.ts` | Stopped D4 | `("3\n", "", 0)` | Stopped D4 | Stopped D4 | Stopped D4 |
| main | `subclass.ts` | Stopped D5 | `("derived\n", "", 0)` | Stopped D5 | Stopped D5 | Stopped D5 |

| Diagnostic | Exact compiler stderr (JSON-escaped, including final newline) |
|---|---|
| D1 | `"adamic: /tmp/p/class_method.ts:9:20: Adamic 0.1 refuses a method read as a value (add would lose its object, and this with it); call it in an arrow that keeps the object: (by) => counter.add(by) (unbound-method)\n"` |
| D2 | `"adamic: /tmp/p/interface_view.ts:12:20: Adamic 0.1 refuses a method read as a value (add would lose its object, and this with it); call it in an arrow that keeps the object: (by) => adder.add(by) (unbound-method)\n"` |
| D3 | `"adamic: /tmp/p/parenthesized.ts:2:21: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => box.read() (unbound-method)\n"` |
| D4 | `"adamic: /tmp/p/double.ts:2:20: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => box.read() (unbound-method)\n"` |
| D5 | `"adamic: /tmp/p/subclass.ts:12:13: Adamic 0.1 refuses a method read as a value (label would lose its object, and this with it); call it in an arrow that keeps the object: () => item.label() (unbound-method)\n"` |

At head, no program that lowers prints anything different from Node or crashes. `class_method.ts`, `interface_view.ts`, `parenthesized.ts`, and `subclass.ts` match Node on all three backends. `double.ts` does not lower at either revision. Main refuses all five programs.

Committed only `proof/lower-bang/report.md` on `proof/lower-bang-report`, created from the verified head, and pushed only that branch to origin.
