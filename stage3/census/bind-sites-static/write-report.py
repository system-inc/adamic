#!/usr/bin/env python3
"""Render recorded API evidence; this script does no source/type/flow discovery."""
import collections, gzip, hashlib, json, pathlib, sys
here=pathlib.Path(__file__).resolve().parent
source, pristine, rule, adapted=map(pathlib.Path,sys.argv[1:5])
apply_log=pathlib.Path(sys.argv[5])
load=lambda name:json.loads((source/name).read_text())
inv=load('inventory.json'); candidates=load('candidates.json'); upper=load('member-values-upper-bound.json'); flow=load('source-flow.json')
old=json.loads((pristine/'inventory.json').read_text())
assert len(inv['roots'])==79 and len(inv['files'])==79 and not inv['diagnostics']
assert len(inv['sites'])==56 and all(s['kind']=='bind' for s in inv['sites'])
assert len(old['sites'])==56
oldgroups=collections.defaultdict(list)
for s in old['sites']:oldgroups[(s['file'],s['member'])].append(s)
used=collections.Counter(); origins=[]
for s in inv['sites']:
 k=(s['file'],s['member']);origins.append(oldgroups[k][used[k]]);used[k]+=1
assert all(used[k]==len(v) for k,v in oldgroups.items())
status=[]
for s,c in zip(inv['sites'],candidates):
 assert (s['file'],s['line'])==(c['file'],c['line'])
 if s['member']=='fn':status.append('H')
 elif any(f['rest'] or f['readsArguments'] for f in c['typed']):status.append('P')
 elif c['extra']:status.append('a*')
 else:status.append('c*')
counts=collections.Counter(status)
assert counts=={'H':1,'P':21,'a*':15,'c*':19}
# IDs are report-local, deterministically assigned to recorded API objects.
values={v['key']:v for s in upper['sites'] for v in s['values']}
for s in flow['sites']:
 for v in s['values']:values.setdefault(v['key'],v)
vid={k:('W' if v['kind']=='bound-wrapper' else 'F')+str(i+1).zfill(3) for i,(k,v) in enumerate(sorted(values.items()))}
finfo={f['key']:f for f in inv['functions']}
site_key=lambda s:s['file']+':'+str(s['line'])+':'+str(s['column'])
types={}; typeids=[]; groups={};groupids=[]
for s,u in zip(inv['sites'],upper['sites']):
 t=s['typeText'];types.setdefault(t,'T'+str(len(types)+1).zfill(2));typeids.append(types[t])
 name=s['member'].split('.')[-1];groups.setdefault(name,('I'+str(len(groups)+1).zfill(2),u));groupids.append(groups[name][0])
fmt=lambda x:'`'+str(x).replace('|','\\|').replace('\n',' ')+'`'
ref=lambda s:fmt(s['file']+':'+str(s['line'])+':'+str(s['column']))
q=[]
def p(s=''):q.append(s)
def quote(file,a,b):
 lines=(rule/file).read_text().splitlines();p(f'`{file}:{a}–{b}` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:');p();p('```go')
 for i in range(a,b+1):p(f'{i}: {lines[i-1]}'.rstrip())
 p('```');p()
p('**Partial static projection: (a) 15 candidate new count stops; (b) 0 demonstrated known-reach overflows; (c) 19 source-pool count-clean sites. Another 21 have an older rest-bearing candidate and 1 is the unused helper body.**')
p()
p('The asterisk on `a*` and `c*` below is essential: these are conditional source-level classifications, **not a certified census of the head’s completed IR `closureRecords`**. The exact rule, root closure, all site identities, member types and source parameter counts are measured. The source implementation/flow sets are explicitly incomplete/conservative; generated and specialized IR records have not been exhaustively reconstructed. No claim that tsc reaches this pass, newly compiles, or actually stops at 15 sites is made. This is a measurement/report only; compiler and adapters are unchanged.')
p()
p('**Scope correction:** the requested pins produce **56 `.bind` calls and zero `maybeBind` calls**, not 12 and 31. Pristine v6.0.3 has 10 `.bind` calls (including the helper) and 46 `maybeBind` calls. Main’s 32, 95h and 99x adapters rewrite all 46 calls. There are zero such sites in `binder.ts`. All 56 actual sites are retained below, with their pristine origins. No requested site was silently omitted to fit the supplied counts.')
p()
p('## Exact head rule')
p()
p('`M` in this rule is the length of the generated bound invocation’s `Arguments`, not JavaScript `.length`, not the number of required parameters, and not the number of arguments at the bind call. The wrapper sends the fixed bind prefix plus the target signature’s remaining declared slots. Thus optional/defaulted slots count. If the fixed prefix itself exceeds the declared slots, its size controls the invocation count. Every actual non-helper site here binds only the receiver, so its invocation count equals the target signature parameter count tabulated as M.')
p()
quote('internal/lower/function_bind.go',113,128)
quote('internal/lower/function_bind.go',363,381)
p('The count excludes an IR method’s inserted leading receiver; captured environment cells are not parameters. A written lone `this` parameter is conservatively counted. Source N below counts written parameter declarations, including optional/defaulted/destructured/rest declarations; class insertion is kept separate.')
p()
p('A known reach requires the allocation graph to report no unknown frontier and every reached site to be one registered by this query as a `MakeClosure`, `BoundClosure`, undefined, or null. Only the first two contribute functions; null/undefined contribute none. Any other origin, untracked producer/argument/result, graph-owned object/array allocation or unknown projection makes the reach unknown. There is no “use the known subset” fallback.')
p()
quote('internal/lower/function_bind.go',405,445)
p('The unknown fallback is **all recorded lowered function implementations whose recorded checker type is assignable to the bound checker type**, irrespective of member name, host identity, assignment into that member, execution liveness or binding at this site. It is neither “every function with this member name” nor “all functions of arity M”. It includes recorded source/module functions, closures, bound wrappers, specializations and generated callable adapters when their records are present. Class method-table functions under `BoundName` are added separately and are never narrowed by the reach result.')
p()
quote('internal/lower/arguments_length.go',139,177)
quote('internal/ir/call_targets.go',174,203)
p('Existing argument-count/rest refusal takes precedence across the entire type/name candidate set. Only if it did not find a refusal does the extra-parameter check run. It asks reach only when some type-wide candidate has N > M; known reach replaces the typed count set, unknown reach retains the whole typed set. Named methods remain in either case. Any counted N > M stops. A known reach containing such a function therefore also stops; categories (a) and the question’s literal (b) overlap. This report uses stop-first counting and reserves (b) for a demonstrated overflow escaping the new guard. None is established.')
p()
quote('internal/lower/function_bind.go',459,504)
p('Projection certificates are globally conservative. Any `CallClosure`, `ObjectCall`, generator/iterator step or yield anywhere in Main or any function body sets `opaqueCalls`; every property/element projection then records an unknown reason, even if its literal field can be identified. Opaque field stores, mutation, spread, absent/unmodelled fields, recursive projection and dynamic element keys add other unknown reasons.')
p()
quote('internal/lower/shape_flow.go',267,279)
quote('internal/lower/shape_flow.go',319,335)
quote('internal/lower/shape_flow.go',359,364)
p('**All 55 non-helper sites use the direct same-identifier method-bind path. If their count reach is queried, it is unknown in this head.** The method path builds a `BoundMethod` whose factory captures the runtime lookup, rather than an ordinary IR `Call` with the placeholder target argument. The shared graph’s call/parameter collection has no `BoundMethod` argument edge. Its factory’s first `bound_value` parameter therefore has no tracked producer. `reachingAllocations` reports that as unknown. Even the three resolutionCache members whose source object literals are statically identifiable do not get a known head certificate. Opaque projection calls provide an additional barrier when source properties are followed; do not equate the supplementary source-flow result with head reach.')
p()
quote('internal/lower/function_bind.go',62,74)
quote('internal/lower/function_bind.go',137,156)
quote('internal/lower/view_flow.go',123,143)
quote('internal/lower/view_flow.go',153,185)
quote('internal/lower/shape_flow.go',123,139)
p('The pass returns the first recorded error (`function_bind.go:508–512`), and runs only after ordinary lowering, cycle discovery and count-target construction (`lower.go:215–220`). Static site rows are not an observed error sequence.')
p()
p('## maybeBind and the adapted program')
p()
quote('internal/lower/function_bind.go',28,35)
p('The retained helper is `src/compiler/core.ts:1519–1521`, pristine `:1522–1524`:')
p()
p('```typescript\nexport function maybeBind<T, A extends any[], R>(obj: T, fn: ((this: T, ...args: A) => R) | undefined): ((...args: A) => R) | undefined {\n    return fn?.bind(obj);\n}\n```')
p()
p('Adamic monomorphizes generic direct calls: `generic.go:24–49` infers each call’s concrete parameter/result mapping, `:80–114` keys/reuses instances, `:141–168` lowers with the mapper, and `instantiate.go:50–56` substitutes types through `concrete`. It does not judge the helper once as a monomorphic source function with its written generic constraint. At this exact adapted pin **there are no core-helper callers and hence no caller instantiations to enumerate**. If a caller were retained, `fn?.bind(obj)` has an identifier target and is rejected by the quoted optional-chain rule before parameter counting, regardless of the caller’s instantiation. The rewritten `host.member?.bind(host)` sites have a property target and pass that particular syntactic check; they are judged independently with the member’s concrete type. The helper row is `H`, not a fabricated per-caller count verdict.')
p()
p('The API program uses all 79 recursively enumerated `.ts` roots under `src/compiler`, the pinned `@types/node@25.3.3` ambient index and Adamic’s strict ES2024/bundler type options. It reports **zero TypeScript diagnostics**. All 79 loaded non-declaration sources are within `src/compiler`; **no services, server or harness source is loaded**. System’s source implementation in `sys.ts` is loaded and inventoried. Ambient Node declarations are not source implementations and have no N/body to invent.')
p()
p('## One row per actual adapted site')
p()
p('`U` = the head reach would be unknown if queried; `—` = earlier rest refusal means the reach query is not made, or no typed overflow requires one. `U/—` states both the structural result and that query gating. `a*` = source-pool count-stop candidate; `c*` = source-pool count-clean; `P` = an existing rest-bearing source candidate takes precedence if recorded; `H` = unused helper body/earlier optional-chain blocker. **P and H are not called clean.** The API source-pool relation is an upper-bound audit, not the completed IR table. Every `a*` has an explicit non-generic-declaration witness (`Debug.fail` N=2, or the overloaded `assertOptionalToken` implementation N=4 for setTimeout); absence of additional generated earlier stops is not certified.')
p()
p('Implementation IDs in column I point to the conservative same-field source audit below. “Source flows” lists values found by the supplementary allocation pass; `?` retains an unknown ingress and **does not mean only those values flow**. Each implementation/wrapper ID gives N and a source citation in the catalogue. The three source-known cache rows still have head reach U/—.')
p()
p('| # | Adapted site and quoted call (pristine origin) | Member type / M | I | Enumerated source flows | Head reach | Verdict |')
p('|---|---|---|---|---|---|---|')
for i,(s,c,u,f,o,st) in enumerate(zip(inv['sites'],candidates,upper['sites'],flow['sites'],origins,status)):
 vals=', '.join(vid[v['key']] for v in f['values']) or 'none enumerated'
 if f['unknown']:vals+='; ?'
 h='not instantiated' if st=='H' else 'U (queried)' if st=='a*' else 'U/—'
 origin=f"{o['file']}:{o['line']} ({o['kind']})"
 p(f"| B{i+1:02} | {ref(s)} {fmt(s['text'])}; pristine {fmt(origin)} | {typeids[i]}, M={','.join(map(str,s['M']))} | {groupids[i]} | {vals} | {h} | {st} |")
p()
p('## Member types')
p()
p('Types below are `getNonNullableType(getTypeAtLocation(member))` and call signatures from the pinned checker. The declaration citations are the original member contracts, not implementations. Optional membership is removed only for reading the callable signature. Full declaration text and resolved bind signatures are retained in `evidence/inventory.json.gz`.')
p()
for t,tid in types.items():
 s=inv['sites'][typeids.index(tid)]
 p(f"- **{tid}**, M={','.join(map(str,s['M']))}: {fmt(t)}. Contract: "+', '.join(ref(d) for d in s.get('declarations') or [])+'.')
p()
p('## Conservative implementation catalogue (incomplete flow coverage)')
p()
p('This catalogue joins definitions/writes of the same property name across receivers, then follows source symbols, initializers, known call returns and argument edges. It deliberately includes unrelated same-name implementations, e.g. six-parameter `WatchFactory` adapters do not thereby become System’s four-parameter watcher. It also retains unresolved/dynamic/frontier cases. **It is not an exhaustive set of values proven to flow into each particular bound member.** That requested part remains unfinished. The narrower allocation pass is an aid, not a replacement completeness proof: it is context/flow insensitive and filters object allocations by host-bearing properties within depth two. Its fixed point took '+str(flow['steps'])+' steps, with '+str(flow['slots'])+' slots and '+str(flow['callEdges'])+' call edges. All non-helper host inputs remain open except the three local resolutionCache object members.')
p()
p('No loaded class method has any of the bound member names (`candidates.json.gz`, class-only method-table audit). All enumerated host implementations are source functions, object-literal closures or wrappers. No harness/class-host body is silently imported from TypeScript’s tests. For properties with only wrapper cycles or no in-program definition, outside callers may supply a function; the source program does not identify its N.')
p()
for name,(gid,u) in groups.items():
 p(f"- **{gid} `{name}`**: "+(', '.join(vid[v['key']]+f" (N={v['N']})" for v in u['values']) or 'no source implementation enumerated')+'.')
 if u['frontiers']:p('  Audit frontiers: '+ '; '.join(fmt(s) for s in u['frontiers'])+'.')
p()
p('### Implementation and wrapper source declarations')
p()
for k,v in sorted(values.items()):
 ident=vid[k]
 if v['kind']=='bound-wrapper':p(f"- **{ident}**, N={v['N']}: {fmt(k)}, generated by the source bind of {fmt(v.get('target',''))}. This is the adapter’s declared slot count, not the original target’s N.")
 else:
  f=finfo.get(k)
  p(f"- **{ident}**, N={v['N']}: {fmt(k)} {fmt(f['header'] if f else v.get('signature',v.get('name','')))}"+ ('; rest parameter present.' if v.get('rest') else '.'))
p()
p('## Source-pool stop evidence and limits')
p()
p('For every site, `candidates.json.gz` retains **all** source callable-body declarations assignable under the stock checker to its target type, all N > M candidates, argument readers/rest flags and class-name candidates. Constructors/accessors are excluded from this callable pool. This pool is not renamed as the head’s `FunctionTypeTargets`: it can include uninstantiated generic/nested bodies, and does not enumerate every generated/specialized IR record. The positive witnesses below establish why the new fail-closed count predicate matters; they do not certify the entire IR candidate universe or the absence of an earlier generated stop.')
p()
p('| Site | M | Explicit source witness | N |')
p('|---|---:|---|---:|')
for i,(s,c,st) in enumerate(zip(inv['sites'],candidates,status)):
 if st!='a*':continue
 w=next((f for f in c['extra'] if f['name']=='fail'),c['extra'][0])
 p(f"| B{i+1:02} | {s['M'][0]} | {fmt(w['key'])} {fmt(w['name'])} | {w['N']} |")
p()
p('`Debug.fail` is declared at adapted `src/compiler/debug.ts:174` as `fail(message?: string, stackCrawlMark?: AnyFunction): never` (N=2); TypeScript permits it through the zero-/one-slot target signatures above. `assertOptionalToken` at `debug.ts:321–324` has overloads and a four-parameter implementation; the checker relates its overload type to the three-slot setTimeout target. Neither witness need ever be assigned to the bound member: that is exactly what the unknown-reach fallback counts.')
p()
p('For string-returning and void targets, unrelated `combinePaths(path: string, ...paths: (string | undefined)[]): string` (`path.ts:581`) and `resolvePath` (`:608`) are rest-bearing assignable candidates. The cached writeFile closure (`program.ts:583`) and generic predicate closure (`core.ts:2489`) account for the other recorded source rest candidates. A rest flag inferred from an uninstantiated generic body is not a certified `RestElement` in the missing completed IR. P is retained as an earlier-stop bucket rather than inflated into “new count stops”.')
p()
p('**Stopping reason / remaining coverage:** exact per-site, exhaustive implementation-flow sets and a complete reconstruction of specialized/generated IR closure records are not established by these API analyses. In particular, a field-name join is not a receiver-flow proof and a stock-checker generic signature is not a record of every concrete Adamic specialization. Publishing the 15/19 projection as an unconditional 56-site a/b/c verdict would hide that gap. There are no observed native tsc wrong answers or new whole-program stops in this report. The rule and all 56 sites/types are complete; exhaustive flow certification is partial.')
p()
if (here/'USES.md').exists():
 p((here/'USES.md').read_text().rstrip());p()
p('## Reproduction and provenance')
p()
p('Pins: rule `ba619e28774130e24331c3026a938cd6f3cd6a34`; adapters/base `55c375d9d245b580c010c3e0b5659f2849f24fb6`; TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`; cohere `5360753e19e4417675a90e92a1823bea66b13918`; cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Both Adamic worktrees were fresh detached worktrees; the report branch was subsequently created on the main worktree. Recursive depth-one submodule initialization and `npm ci` in stage3/api succeeded.')
p()
p('```sh\ngit fetch origin ba619e28774130e24331c3026a938cd6f3cd6a34 55c375d9d245b580c010c3e0b5659f2849f24fb6\ngit worktree add --detach "$SCRATCH/rule" ba619e28774130e24331c3026a938cd6f3cd6a34\ngit worktree add --detach "$SCRATCH/main" 55c375d9d245b580c010c3e0b5659f2849f24fb6\ngit -C "$SCRATCH/rule" submodule update --init --recursive --depth 1\ngit -C "$SCRATCH/main" submodule update --init --recursive --depth 1\ngit clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$SCRATCH/TypeScript"\ngit -C "$SCRATCH/TypeScript" rev-parse HEAD\nexport ADAMIC_TYPESCRIPT_SOURCE="$SCRATCH/TypeScript"\n(cd "$SCRATCH/main/stage3/api" && npm ci)\n(cd "$SCRATCH/main" && stage3/apply.sh "$SCRATCH/adapted")\n# Run the scripts from this report commit, whose API dependencies are installed.\nnode stage3/census/bind-sites-static/run.cjs "$SCRATCH/adapted" "$SCRATCH/evidence"\n```')
p()
p('For the pristine-origin map, run `measure.cjs "$SCRATCH/TypeScript" "$SCRATCH/pristine-evidence"`; the matching uses API-discovered file/member identities and occurrence order, with all 56 origins required to match. It does not use regex to discover sites, types or flow. Python only renders/compresses already-recorded API results and quotes the rule source. No census run or production compiler modification is used to bypass earlier failures.')
p()
p('Validation: stock TypeScript 6.0.3; 79/79 roots present; 79/79 loaded implementation files within compiler; 0 diagnostics; 56/56 sites and pristine origins matched; 0 core-helper callers; 0 bound-name class method declarations; fixed-point supplemental flow terminated. `evidence/SHA256SUMS` hashes the committed compressed API evidence, original-site inventory, apply log and patch-set receipt. The per-source SHA-256s are recorded in the inventory. The applied patch-set total is 118 files, 7,640 added and 6,796 removed lines.')
p()
(here/'REPORT.md').write_text('\n'.join(q).rstrip()+'\n')
e=here/'evidence';e.mkdir(exist_ok=True)
for name in ['inventory.json','candidates.json','member-values-upper-bound.json','source-flow.json']:
 (e/(name+'.gz')).write_bytes(gzip.compress((source/name).read_bytes(),mtime=0))
(e/'pristine-inventory.json.gz').write_bytes(gzip.compress((pristine/'inventory.json').read_bytes(),mtime=0))
(e/'apply.log.gz').write_bytes(gzip.compress(apply_log.read_bytes(),mtime=0))
(e/'patch-set.md').write_bytes((adapted/'patch-set.md').read_bytes())
summary={'status':'partial source-level projection; not completed IR census','counts':dict(counts),'roots':len(inv['roots']),'source_files':len(inv['files']),'sites':len(inv['sites']),'maybeBind_calls':0,'diagnostics':len(inv['diagnostics']),'source_flow_steps':flow['steps'],'not_covered':['exhaustive exact receiver-sensitive function-flow sets','complete generated/specialized IR closureRecords','actual whole-program lowering outcome or native tsc execution']}
(e/'SUMMARY.json').write_text(json.dumps(summary,indent=2)+'\n')
(e/'SHA256SUMS').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+f.name+'\n' for f in sorted(e.iterdir()) if f.name!='SHA256SUMS'))
print(json.dumps(summary,indent=2))
