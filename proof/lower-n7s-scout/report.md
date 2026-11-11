The gap is ordinary captured-Union admission, not recognition of Session or of the string literal. On the requested main revision, both a host-Session union and a program-class union lower as uncaptured `ir.Union` locals, but fail when two arrows capture the binding. The compiler already admits a placeholder Union capture using the same cell payload. There are **two matching captured let declarations** in pinned upstream `src/compiler`, counting arrays as object types.

**Scope and completed setup.** Examined Adamic `16429a30452d027809f0ea85bc395479423e493e`, checked out detached after `git fetch origin main`, then ran `git submodule update --init --recursive`. Pinned submodules are cohere `5360753e19e4417675a90e92a1823bea66b13918` and cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Upstream Microsoft/TypeScript was fetched and checked out outside the repository at `050880ce59e30b356b686bd3144efe24f875ebc8`, in `/tmp/lower-n7s/upstream`. Created `/tmp/adamic-gate` outside the repository before Go ran. Copied `stage3/api/package.json` and its lockfile into `/tmp/lower-n7s/api`, ran `npm ci --ignore-scripts --prefix /tmp/lower-n7s/api`, and verified `tsc --version` is `Version 6.0.3`. No repository package install or source change was made.

The setup/probe environment was:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lower-n7s/upstream
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export PATH=/tmp/lower-n7s/api/node_modules/.bin:$PATH
```

All source citations below refer to that Adamic revision unless prefixed `upstream/`. This is a reading and small-lowering-probe report, not a whole-tsc lowering or backend execution result. Only this report is authored in the repository.

**1. What actually decides the stop.** `internal/lower/expression.go:1335–1336` is diagnostic vocabulary, not the capture guard:

```go
case ir.Union:
    return "union of differently held members"
```

The representation decision is `representation` in `internal/lower/expression.go:56–206`. In particular, `:181–195` recursively represents each non-nullish union member, compares their **ir.Type values**, and returns Union if two differ:

```go
if shared != 0 && memberType != shared {
    mixed = true
}
...
if mixed {
    // Members held differently (string | number) are one Union, which holds undefined too.
    return ir.Union, true
}
```

Thus “differently held” means different runtime IR representations, not simply different TypeScript types and not simply counted versus uncounted. Two ordinary class-instance types both map to Object, and two string literals both map to String. Object versus String is different even though both are counted references; Number versus Boolean is different even though neither is a counted reference.

| Proven member/category | Representation and deciding source |
| --- | --- |
| Numeric types/literals; booleans; strings/literals | `ir.Number`, `ir.Boolean`, `ir.String`; `expression.go:126–131` |
| Ordinary non-callable objects, including Session and the probe's class | `ir.Object`; `expression.go:143–144` |
| Arrays and supported RegExp result arrays | `ir.Array`; `expression.go:132–133`; NodeArray and Buffer special cases at `:58–59`, `:79–80` |
| Map/ReadonlyMap/Set/ReadonlySet; Promise; callable objects | `ir.Map`, `ir.Promise`, `ir.Closure`; `expression.go:134–147`; library Function at `:64–66` |
| Representable indexed dictionaries/finite partial records | `ir.Record`; `expression.go:85–90` |
| Typed arrays | Their individual typed-array `ir.Type`, so e.g. Uint8Array and Int32Array differ; `expression.go:61–62` |
| `unknown`, `object`, or structural object types that admit primitives | Already `ir.Union`; `expression.go:69–76`, `:139–142` |
| Null/undefined/void individually | `ir.Object`; `expression.go:124–125`; nullish members are normally skipped during the shared-member comparison at `:170–174` |

There are important qualifications to that table. Null-containing unions have an earlier boxed-representation decision (`expression.go:149–166`): null plus undefined, a non-reference arm, String, Union, or Record forces Union there. Otherwise null/undefined can use a reference's missing-pointer protocol. If all remaining members share a scalar representation and undefined is included, `expression.go:197–204` uses `ir.Maybe(shared)` (`MaybeNumber` or `MaybeBoolean`), not Union. Weak unions have their own handling at `:176–191`; type parameters/intersections are resolved or rejected at `:105–121`. An unrepresentable member returns unknown/failure at `:181–184`, rather than automatically becoming Union. `ir.Type.IsReference` includes String, Object, Array, Map, Record, Closure, Union, Weak, Promise, Generator, and typed arrays (`internal/ir/ir.go:354–356`).

For the requested declaration, Session is Object, `"stopping"` is String, and undefined is skipped in the shared-member comparison: Object != String, hence **Union**. Its intended meaning is already documented at `internal/ir/ir.go:296–299`: “one counted reference, which says at runtime which member it is.” Numbers are boxed, booleans have constant boxes, and undefined is a null reference. Strings and objects can retain their existing heap representations; `internal/native/union.go:25–34` boxes scalars and casts reference arms to `adamic_heap *`.

The actual capture refusal is `internal/lower/locals.go:197–208`. A local lookup calls `touch(local)` at `:200`; `touch` skips globals and same-function locals (`:222–224`), sets `Captured = true` (`:225`), and puts the local into intervening closures' Environment lists (`:237–240`). Its comment at `:214–216` says a captured local “moves into a cell,” shared by reference. The guard then reads:

```go
if declared := l.result.Locals[local]; declared.Captured && slotless(declared.Type) && !(declared.Placeholder != "" && declared.Type == ir.Union) && l.unlowerable == nil {
    // A placeholder union is one counted heap reference, which a cell already holds.
    // Other slotless shapes retain their existing capture refusal.
    l.unlowerable = l.notYet(identifier, "a "+typeName(declared.Type)+" variable a function value captures")
}
```

`slotless` returns true for **MaybeBoolean and Union**, exactly (`expression.go:1362–1369`). Its comment describes contexts that have not arranged packing/boxing, including cells. Immediately afterward, `boxedSlot` explicitly admits fitted Union as “one counted reference” (`:1372–1375`). This is therefore a conservative context-admission rule, not evidence that Union cannot physically fit a cell. The placeholder exception already acknowledges that distinction. The diagnostic is deferred through `l.unlowerable` because lookup cannot return an error, then returned by `Lower` at `internal/lower/lower.go:158–159`. It is `*lower.NotYet`, not the permanent `*lower.Refused`; their error formats differ (`internal/lower/diagnostics.go:10–34`).

An uncaptured local never satisfies the Captured condition. Its declared representation is assigned by `declareLocal` (`locals.go:104–143`), initialized with `fit` (`:78–92`), and assigned with `fit` (`internal/lower/assignments.go:111`). `fit` emits `ir.Box` when a non-Union value goes into Union (`expression.go:878–880`). Truthiness already goes through `condition` → `censusCondition` (`internal/lower/control.go:98–107`, `internal/lower/census_small.go:429–435`); strict equality fits both operands to Union if either is Union and compares them (`expression.go:1098–1117`). Reads keep the stored Union for equality, and use checked narrowing for appropriate member-consuming reads (`locals.go:272–284`, `:344–363`; `internal/lower/boxed_scalar_fields.go:45–55`).

The upstream motivating declaration is `upstream/src/compiler/sys.ts:1481:13`. Its enable function tests it at `:1646:17`; disable tests it against `"stopping"` at `:1694`, clears it at `:1709`, and writes the literal at `:1713`. There is also a capture in `cpuProfilingEnabled` at `:1592:42`. These are multiple uses of one binding, not multiple declaration sites.

**2. Main-revision probes.** All fixtures lived in `/tmp/lower-n7s/probes`. The driver lived in an external module `/tmp/lower-n7s/driver`, named `github.com/system-inc/adamic/scout` so imports obey Go's internal-package visibility. It required Adamic v0.0.0 replaced by `/workspace/adamic`, and copied the repository's go.mod replacement/dependency entries with relative cohere paths made absolute. Go could therefore use the repository's actual packages without adding a command or test file there.

The entry is exactly the one used by `lowerSourceAs` at `internal/lower/lower_test.go:23–29`: `load.Load([]string{file})`, then `lower.Lower(context.Background(), program)`. The essential driver, including the representation/environment inspection, was:

```go
package main
import (
 "context"
 "fmt"
 "os"
 "github.com/system-inc/adamic/internal/load"
 "github.com/system-inc/adamic/internal/lower"
 "github.com/system-inc/adamic/internal/ir"
)
func main() {
 for _, file := range os.Args[1:] {
  fmt.Println("PROBE",file)
  p,e:=load.Load([]string{file}); if e!=nil {fmt.Println("LOAD",e);continue}
  result,e:=lower.Lower(context.Background(),p);if e!=nil {fmt.Printf("LOWER %T %v\n",e,e);continue}
  fmt.Printf("LOWER OK Union=%d Object=%d String=%d\n",ir.Union,ir.Object,ir.String)
  for i,l:=range result.Locals {if l.Name=="x" {fmt.Printf("LOCAL %d %+v\n",i,l)}}
  for i,f:=range result.Functions {if f.Closure {fmt.Printf("CLOSURE %d env=%v\n",i,f.Environment)}}
  fmt.Printf("STRINGS %q\n",result.Strings)
 }
}
```

For `host-*.ts`, the first line was `import type { Session as Obj } from "node:inspector";`. For `class-*.ts`, it was `class Obj { value: number = 1; }`. No Session constructor, host method, or host implementation was needed: the object arm came from a typed parameter, isolating the representation/capture question.

The uncaptured body, identical for both prefixes, was:

```ts
function probe(obj: Obj, tag: number): boolean {
  let x: Obj | "stopping" | undefined;
  if (tag === 0) { x = obj; }
  else if (tag === 1) { x = "stopping"; }
  else { x = undefined; }
  if (x) { return x === "stopping"; }
  return false;
}
```

The two-arrow body, also identical for both prefixes, was:

```ts
function probe(obj: Obj, tag: number): boolean {
  let x: Obj | "stopping" | undefined;
  const read = (): boolean => {
    if (x) { return x === "stopping"; }
    return false;
  };
  const write = (): void => {
    if (tag === 0) { x = obj; }
    else if (tag === 1) { x = "stopping"; }
    else { x = undefined; }
  };
  write();
  return read();
}
```

Each writer covers all three members; the unknown tag keeps all members possible at the reads. In the captured fixtures the reader is deliberately declared first, placing the first capture at `if (x)`. These bindings are inside a function, so they exercise real cells rather than top-level globals (which `touch` exempts).

| Fixture | Lower result | Observed representation |
| --- | --- | --- |
| (a) `host-local.ts` | Lowers | Local 2, `Name:x Type:10 Captured:false`, i.e. `ir.Union`; no captured cell |
| (b) `host-captured.ts` | NotYet at 5:9 | No successful IR returned; no admitted cell. Declaration's representation is Union by the rule above |
| (c) `class-local.ts` | Lowers | Local 2, `Name:x Type:10 Captured:false`, i.e. `ir.Union`; no captured cell |
| (c) `class-captured.ts` | NotYet at 5:9 | Same capture guard and no successful IR returned |
| Extra control: `class-placeholder-control.ts` | Lowers | Local 2, `Type:10 Captured:true Placeholder:x Uninitialized:true`; both arrows share local 2 in their environments |

Both successful ordinary probes intern only `"stopping"` in `Program.Strings`. The driver confirms `ir.Union == 10`, `ir.Object == 4`, and `ir.String == 3`. There is no separate Cell `ir.Type`: `ir.Local.Captured` chooses cell storage while `ir.Local.Type` specifies its payload (`internal/ir/ir.go:359`, `:386–387`). The placeholder control changes only the declaration to `let x: Obj | "stopping" | undefined = undefined!;`; its read-arrow environment is `[2]` and its write-arrow environment is `[1 2 0]` (tag, x, obj). That proves existing lowering can produce a captured **Union payload**, while preserving placeholder-specific semantics; it is not proposed as a source workaround.

Exact ordinary-capture errors:

```text
/tmp/lower-n7s/probes/host-captured.ts:5:9: stage 0 can't lower a union of differently held members variable a function value captures yet
/tmp/lower-n7s/probes/class-captured.ts:5:9: stage 0 can't lower a union of differently held members variable a function value captures yet
```

The driver printed `LOWER *lower.NotYet` before each error. Every fixture separately passed this strict tsc 6.0.3 check:

```sh
tsc --noEmit --strict --skipLibCheck --types node   --typeRoots /tmp/lower-n7s/api/node_modules/@types   --target es2020 --module nodenext "$probe"
```

Ran the driver with all five absolute fixture paths using `go run -mod=mod . ...` from its external module. No compiler rule, registration, or repository source overlay was changed. These are lowering results, not claims that the fixtures were emitted or run.

**3. Reach in pinned upstream.** Count: **2 distinct captured let declarations**, both explicitly annotated. Arrays are object types in TypeScript, so `string[]` counts as the object arm. If “object” is restricted to non-array objects, the count is **1**, activeSession.

| Declaration site | Declared union | Capturing function/use |
| --- | --- | --- |
| `src/compiler/sys.ts:1481:13` | `import("inspector").Session \| "stopping" \| undefined` | `enableCPUProfiler`, `disableCPUProfiler`, nested callbacks, and the `cpuProfilingEnabled` arrow; e.g. `sys.ts:1646:17`, `:1694:17`, `:1592:42` |
| `src/compiler/commandLineParser.ts:3577:9` | `string \| string[] \| undefined` | `onPropertySet` (declared at `:3597`), writes the binding at `:3632:17` |

Method: a TypeScript **6.0.3 compiler-API script**, not grep and not typescript-go. It read the upstream compiler tsconfig (strict checking), resolved declaration types and identifier symbols, selected `let` declarations whose declared union includes an object/object-intersection arm, a StringLike arm, and undefined/null, then counted a symbol if a reference occurs inside a nested function-like node beneath the declaring function. It also checked inferred symbol types, not only written annotations. All 77 on-disk `.ts` source files under `src/compiler`, including subdirectories, were scanned. There were 8 type-shape candidates overall (5 explicitly annotated), and only the two above had captures.

For audit, the six uncaptured candidates were `factory/nodeFactory.ts:3531` (cooked), `parser.ts:9616`, `:9721`, `:9732` (comment), `commandLineParser.ts:3527` (a different extendedConfigPath), and `binder.ts:2362` (pattern). The Session arm resolved as Session, not any; the array arm resolved as string[].

This is an exact symbol/lexical-capture count for those declarations in the pinned tree, including write-only captures. It counts neither references nor runtime invocations, makes no reachability claim, and does not imply each site's whole enclosing function currently lowers. The script uses function-like scope ancestry and symbol identity; it does not model Adamic's capture analysis or closure escape optimization. The reported matching references were also read in source and are runtime expressions, not type-only references. Both declaring scopes are functions, not globals.

Upstream lacks its generated diagnostic map in the checkout. To improve type resolution without editing it, ran upstream's `scripts/processDiagnosticMessages.mjs` against a copy of `src/compiler/diagnosticMessages.json` in `/tmp/lower-n7s/generated`, and supplied the generated `.ts` through a virtual compiler host at its expected source path. The census had zero syntactic diagnostics. Its semantic diagnostic audit retained two unrelated issues: the virtual generated file is absent from the composite project's root-file list, and `source-map-support` declarations are not installed. Neither concerns these union types; this is not a claim that the entire upstream compiler typechecks with only stage3/api's packages.

The complete final census script follows. It expects the external upstream/api/generated directories described above; paths can be relocated together. Its `captured` field is the reported count and `rows` retains both accepted and rejected candidates:

```js
const ts = require('/tmp/lower-n7s/api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const root = '/tmp/lower-n7s/upstream';
const configPath = root+'/src/compiler/tsconfig.json';
const cfg = ts.readConfigFile(configPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(cfg.config, ts.sys, path.dirname(configPath));
parsed.options.typeRoots = ['/tmp/lower-n7s/api/node_modules/@types'];
const virtual=root+'/src/compiler/diagnosticInformationMap.generated.ts';
const host=ts.createCompilerHost(parsed.options), read=host.readFile.bind(host), exists=host.fileExists.bind(host);
host.fileExists=f=>f===virtual||exists(f);
host.readFile=f=>f===virtual?fs.readFileSync('/tmp/lower-n7s/generated/diagnosticInformationMap.generated.ts','utf8'):read(f);
const p = ts.createProgram(parsed.fileNames, parsed.options,host), c=p.getTypeChecker();
const files=p.getSourceFiles().filter(s=>s.fileName.startsWith(root+'/src/compiler/') && !s.isDeclarationFile && s.fileName!==virtual);
const owner=n=>{for(let a=n.parent;a;a=a.parent)if(ts.isFunctionLike(a))return a;return n.getSourceFile()};
const where=n=>{const s=n.getSourceFile(),lc=s.getLineAndCharacterOfPosition(n.getStart());return path.relative(root,s.fileName)+':'+(lc.line+1)+':'+(lc.character+1)};
let candidates=new Map();
for(const s of files){function visit(n){if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&ts.isVariableDeclarationList(n.parent)&&(n.parent.flags&ts.NodeFlags.Let)){
const t=n.type?c.getTypeFromTypeNode(n.type):c.getTypeOfSymbolAtLocation(c.getSymbolAtLocation(n.name),n.name),m=t.isUnion()?t.types:[];
if(m.some(x=>x.flags&(ts.TypeFlags.Object|ts.TypeFlags.NonPrimitive) || (x.isIntersection() && x.types.some(y=>y.flags&(ts.TypeFlags.Object|ts.TypeFlags.NonPrimitive))))&&m.some(x=>x.flags&ts.TypeFlags.StringLike)&&m.some(x=>x.flags&(ts.TypeFlags.Undefined|ts.TypeFlags.Null)))candidates.set(c.getSymbolAtLocation(n.name),{node:n,owner:owner(n),type:c.typeToString(t),refs:[]});
}ts.forEachChild(n,visit)}visit(s)}
for(const s of files){function visit(n){if(ts.isIdentifier(n)){const row=candidates.get(c.getSymbolAtLocation(n));if(row&&n!==row.node.name&&owner(n)!==row.owner){let a=n.parent,inside=false;while(a){if(a===row.owner){inside=true;break}a=a.parent}if(inside)row.refs.push(where(n));}}ts.forEachChild(n,visit)}visit(s)}
const semantic=p.getSemanticDiagnostics(); console.error(JSON.stringify({semanticDiagnostics:semantic.length,first:semantic.slice(0,8).map(d=>({where:d.file?where({getSourceFile:()=>d.file,getStart:()=>d.start}):null,message:ts.flattenDiagnosticMessageText(d.messageText," ")}))}));
const rows=[...candidates.values()].map(r=>({site:where(r.node.name),name:r.node.name.text,type:r.type,explicit:!!r.node.type,global:ts.isSourceFile(r.owner),captures:[...new Set(r.refs)],decl:r.node.getText()}));
console.log(JSON.stringify({files:files.length,diagnostics:p.getSyntacticDiagnostics().length,candidates:rows.length,captured:rows.filter(r=>r.captures.length).length,functionLocalCaptured:rows.filter(r=>r.captures.length&&!r.global).length,rows},null,2));
```

**4. Relation to Emit's #qgr7mm4.** Searched the repository and docs with `rg` and tracked files with `git grep`, including:

```sh
git grep -n 'qgr7mm4'
git grep -n -i -E 'qgr7mm4|counted.{0,40}uncounted|uncounted.{0,40}counted'
rg -n -i 'union.*(counted|uncounted)|(counted|uncounted).*union' internal docs notes
```

Also searched for refusals mentioning those terms, and `git log --all --grep=qgr7mm4`. No qgr7mm4 citation was found, and no refusal saying “counted and uncounted union members” (or the reverse) was found. Nearby counted/uncounted hits concerned method-entry dispatch or WASM/count measurements, not this capture diagnostic. Consequently this checkout supplies no issue-specific evidence that #qgr7mm4 landed or that a particular commit implements that issue.

Based on the issue description supplied in the request, this is the **same broad heterogeneous-union storage/ownership family**, but not an identical counted-versus-uncounted discrimination. Session and String are both counted references; their distinct IR types still select Union. Undefined adds the missing-reference state. Conversely, Number versus Boolean would select Union without either arm initially being counted. The capture guard rejects the resulting representation without examining member countedness.

Relevant landed machinery does apply: ordinary Union boxing/fitting (`expression.go:878–880`), reference/scalar conversions (`internal/native/union.go:12–46`), `boxedSlot` (`expression.go:1372–1375`), Map iterator Union slots (`internal/lower/library_map_set.go:229–231`), and placeholder storage/capture (`internal/lower/placeholder.go:49–55`, `locals.go:204–208`). `internal/lower/boxed_union_slots_test.go:10–17` covers admitted boxed-slot paths while retaining a narrower-array refusal. Its introducing history is the landed graph/Map-union change `9a941a49cb2c20ae90130348e5271c0f8cb7f08a`; that is evidence of applicable generic Union work, not proof of a link to qgr7mm4. None of this removes the ordinary captured-Union guard, as both probes show.

**5. Recommended representation road.** Keep the captured local's payload **ir.Union**, with Session/class values as their existing object heap references, `"stopping"` as the existing interned `ir.StringConstant`, and undefined as the missing Union reference. Share one counted reference cell between the closures; do not invent a pointer-sized literal tag or infer a fixed object payload from the first assignment. If null is admitted too, preserve Union's separate null sentinel (`internal/native/union.go:38–40`; `internal/lower/placeholder.go:49–50`). Interning already exists in `internal/lower/locals.go:250–258`, and native StringConstant emission uses that constant's address (`internal/native/emit_expressions.go:91–92`).

The minimal production lowering change belongs in **internal/lower/locals.go:204**, admitting a fitted Union payload for ordinary captures as the placeholder path already does. Reuse `boxedSlot` or introduce a capture-specific admission predicate in **internal/lower/expression.go:1362–1375** if clearer; preserve the existing MaybeBoolean boundary and avoid globally relaxing unrelated slot rules. No new representation selection is needed. Audit, rather than blindly rewrite, `locals.go` declaration/initialization/read paths and **internal/lower/assignments.go:111**: they already fit writes and preserve Union observations. Add focused capture coverage under **internal/lower** (for example `captured_union_test.go`) and backend/oracle witnesses for alternating object/string/undefined writes through both closures, post-call stale narrowing, and cell lifetime/ownership. Do not broaden all Union-slot admissions as part of this fix.

No new IR type or native/JS emission primitive appears necessary. Native already maps Union to `adamic_heap *` and the `reference` slot (`internal/native/emit_values.go:36–37`, `:48–58`), allocates captured cells according to `Local.Type.IsReference()` (`internal/native/emit_locals.go:207–208`, `:223–249`), and retains/releases reference-cell replacements (`:31–64`). JS already stores captured locals in `.value` cells (`internal/javascript/javascript.go:464–489`, `:531–540`) and treats Box as its input value (`:1291–1292`). The expected implementation is lowering admission plus validation on both backends; a scout alone cannot certify every ownership/narrowing path. Removing this stop would not establish that the full CPU-profiler functions lower: inspector host operations remain a separate boundary (see `internal/lower/library_node_inspector_profile_test.go:94–103`, including Session.connect's expected NotYet).
