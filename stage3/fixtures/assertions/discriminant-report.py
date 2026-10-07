import json,collections,pathlib,sys
r=json.load(open(sys.argv[1]))
# Reviewed source spans, discovered by checker/AST rather than spelling searches.
inside={
 'binder.ts':{1688},
 'checker.ts':{14183,14264,18453,25934,25995,33599,33606},
 'factory/nodeFactory.ts':{1335,2332,2376,2936,3001,3071,3187,3255,3370,3396,3778,4278,4310,4552,6049,6075,6121,6136,6328,6337,6352,6361,6395},
 'parser.ts':{4447,9686},
 'program.ts':{3313,3314},
 'utilities.ts':{8473,8491,8499,8509,8511,8523,8525,8535,8537},
}
uncertain={'parser.ts':{7482,8113}}
for w in r['writes']:
 file=w['file'].removeprefix('src/compiler/')
 if w['form'] in ['object initializer','object spread','class field initializer','parameter property']:
  w['construction']='inside';w['explanation']='Direct object creation; this property is initialized before publication of the literal.'
 elif w['line'] in inside.get(file,set()):
  w['construction']='inside';w['explanation']='Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver.'
 elif w['line'] in uncertain.get(file,set()):
  w['construction']='not established';w['explanation']='Modifier returned by parsing; freshness/reuse and earlier publication need an interprocedural parser analysis.'
 else:
  w['construction']='outside local construction'
  p=w['property'];fn=w['function']['name'] if w['function'] else '<top-level>'
  if file=='tsbuildPublic.ts' and w['line']==1921:
   w['explanation']='Retagging a cached build-status record, UpToDate to UpToDateWithUpstreamTypes, so downstream output timestamps are refreshed without rebuilding unchanged declarations. Not a syntax node.'
  elif p=='antecedent':
   w['explanation']='Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis.'
  elif p=='signature':
   w['explanation']='Updating an existing incremental-builder file-info record with a computed declaration signature.'
  elif p=='flags':
   w['explanation']='Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: '+fn+'.'
  elif p in ['externalModuleIndicator','commonJsModuleIndicator']:
   w['explanation']='Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function.'
  elif p=='version':
   w['explanation']='Reusing a watch-host cache entry and changing its version state to false (file presence unknown).'
  else:
   w['explanation']='Updating a receiver supplied to this function; it does not allocate that receiver. Function: '+fn+'.'
# Deliberately remove the observed retag in the mutant. The independent census invariant must fail.
if '--mutant-omit-status-write' in __import__('sys').argv:r['writes']=[w for w in r['writes'] if not(w['file']=='src/compiler/tsbuildPublic.ts' and w['line']==1921 and w['property']=='type')]
if '--mutant-omit-kind-write' in sys.argv:r['writes']=[w for w in r['writes'] if not(w['file']=='src/compiler/utilities.ts' and w['line']==8509 and w['property']=='kind')]
if '--mutant-drop-declaration' in sys.argv:r['writes'][0]['declarations']=[]
if '--mutant-duplicate-write' in sys.argv:r['writes'].append(r['writes'][0])
assert [(w['file'],w['line']) for w in r['writes'] if w['property']=='kind' and w['form']=='=']==[('src/compiler/utilities.ts',8509),('src/compiler/utilities.ts',8523),('src/compiler/utilities.ts',8535)]
assert len([w for w in r['writes'] if w['file']=='src/compiler/tsbuildPublic.ts' and w['line']==1921 and w['property']=='type'])==1,'lost cached-status retag'
assert all(w['declarations'] for w in r['writes']), 'unresolved selected property declaration'
assert len(r['writes'])==len({(w['file'],w['start'],w['property'],w['form']) for w in r['writes']}), 'duplicate write event'
base=pathlib.Path(sys.argv[2]);base.mkdir(parents=True,exist_ok=True)
summary=collections.Counter(w['construction'] for w in r['writes'])
r['summary']={'construction':dict(summary),'population':dict(collections.Counter(w['population'] for w in r['writes']))}
(base/'discriminant-writes.json').write_text(json.dumps(r,separators=(',',':'))+'\n')
def esc(s):return str(s).replace('|','\\|').replace('\n',' ').replace('\r',' ')
def place(x):return f"{x['file']}:{x['line']}:{x['column']}"
def table(rows):
 result=['| Write | Property / form | Written value type | Declaration(s) | Classification / purpose |','| --- | --- | --- | --- | --- |']
 for w in rows:
  ds='; '.join(dict.fromkeys(place(d) for d in w['declarations']))
  result.append('| '+ ' | '.join(map(esc,[place(w),w['property']+' / '+w['form'],w['value_type'],ds,w['construction']+': '+w['explanation']]))+' |')
 return '\n'.join(result)
counts=collections.Counter(w['property'] for w in r['writes'])
props=['kind']+sorted(p for p in counts if p!='kind')
counttable=['| Property | Creation initializer/spread | Other write syntax | Inside | Outside local construction | Not established | Total |','| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for p in props:
 ws=[w for w in r['writes'] if w['property']==p];c=collections.Counter(w['construction'] for w in ws);n=sum(w['form'] in ['object initializer','object spread'] for w in ws)
 counttable.append(f"| {p} | {n} | {len(ws)-n} | {c['inside']} | {c['outside local construction']} | {c['not established']} | {len(ws)} |")
intro=f'''# Discriminant writes in TypeScript 6.0.3

This is a stock-checker census with an explicitly limited construction analysis.
It does **not** establish a complete whole-program escape classification.
Source commit: `050880ce59e30b356b686bd3144efe24f875ebc8`.
The inherited compiler project has **zero stock diagnostics**, and SHA256
checks match all **{r['files']}** original compiler files to the census.
Generated diagnostics and code outside `src/compiler` are excluded.

## Counts and comparison units

**437 resolved checker-candidate write events**, including **79 `kind` writes**:
76 object-literal initializers and 3 constructor assignments, all inside
construction. There are **zero resolved outside-construction `kind` assignments**.
This is not a claim that dynamic-key copying can never write `kind`.

The remaining 358 events comprise **274 other finite-literal discriminant
writes** and **84 broader/partial checker candidates**. The latter are listed
separately rather than silently claiming every checker candidate is a fixed tag.
For example, `CommandLineOption.type` can be a literal **or a Map**, and
control-flow `antecedent` fields can distinguish presence from absence.
`flags` qualifies through `TypeSystemEntity` and `FlowType`; overlapping enum
values mean it is not a unique nominal identity. See the JSON witnesses.

The local construction partition is **{summary['inside']} inside**, **{summary['outside local construction']}
outside local construction**, and **{summary['not established']} not established**.
The distinction between local construction and global escape is essential below.
Assignments/update syntax total 93; creation syntax totals 344
(341 explicit object properties and 3 spread/property events).

'''+ '\n'.join(counttable)+'''

## Exact discovery and counting rule

`discriminant-writes.cjs` uses npm stock TypeScript 6.0.3, the compiler's inherited
tsconfig, and checker types at AST expressions, declarations, type nodes and
call-signature return types, recursively including union constituents and
constraints. No text search discovers properties, unions or writes.

For each resolved union property, stock `getCheckFlags(property)` must contain
`CheckFlags.Discriminant` (HasNonUniformType | HasLiteralType). The ledger saves
the union and every constituent's property type and declaration. Direct top-level
type-parameter property types are excluded from candidate indexing. The finite
subset additionally requires only literal/enum-literal/null/undefined/never
constituents and at least two distinct member property domains. The broader
subset is a **candidate list**, not a reimplementation of stock's private
recursive generic-type predicate or a proof of a globally immutable tag.
All properties named `kind` are counted independently of these predicates,
including pragma flags, comments, module-specifier results and generator blocks.

Each write resolves its receiver property symbol through the checker and root
symbols to declarations. Candidate witnesses must share a declaration; for
explicit accesses the receiver must be stock-assignable to a witness member.
For contextual object literals, declarations are filtered to members admitting
at least one constituent of the initializer's checked value type. This prevents
a Map-valued command-line option initializer from being attributed to the
boolean option's declaration merely because contextual union symbols combine
all declarations. The literal's own declaration is also retained in JSON.
When more than one declaration remains possible, all are reported, not one
chosen by spelling. A property is counted once per syntactic write, even when
many unions witness it. A spread counts once per selected copied property;
it does not count as a mutation of the spread source.

AST forms cover assignment/compound assignment, ++/--, delete, finite literal
bracket keys, destructuring assignment targets, object property/shorthand/method
initializers, spreads, class field initializers and constructor parameter
properties, plus checker-resolved Object.assign and Object.defineProperty.
For compounds the recorded type is the checked result type, rather than the
right operand alone. Counts are source sites, not execution frequencies.

## Exact local construction rule and its limit

An initializer of a new object literal, including spread into that new object,
is inside construction. A function-style allocator constructor's `this` is a
fresh receiver; its assignments count inside only before `this` is returned,
stored externally, or passed to a retaining call. The allocator's Node, Token,
Identifier, Symbol, Type and Signature constructors were inspected. The three
`kind` assignments follow only writes to `this.pos` and `this.end`.

For an ordinary function, an assignment counts inside only at a reviewed span
where all paths reaching it provide a fresh local allocation/clone, and no
preceding operation publishes that receiver. Internal allocating helpers may
return the fresh object to the enclosing construction function; that return
alone is not publication. Property writes on that local and helpers such as
setTextRange/setParent that only populate the fresh receiver are not publication.
Return to the caller, storing into a cache, map, object field or externally
reachable collection, or passing it to an unproved retaining callback ends
construction. The reviewed span list is in `discriminant-report.py`.

Incoming receiver parameters and cache/global/property lookups are outside
**this function's local construction**, even if a caller sometimes invokes the
function while constructing an object. Mixed fresh/existing paths are also
outside the guaranteed construction category. This does not assert that every
such execution is after a global escape. Parser `finishNode` and `withJSDoc`,
for example, are construction helpers whose callers need an interprocedural
freshness proof. I did not do that whole-program proof. The stock checker has
no escape-analysis API; function names and a `const` local do not supply one.
Two modifier-writing parser spans are explicitly not established rather than
assuming that a parsed modifier cannot have been reused or published.

## Definite outside retag

`src/compiler/tsbuildPublic.ts:1921` obtains `status` from
`state.projectStatus.get(nextProjectPath)`. After checking `status.type ===
UpToDateStatusType.UpToDate`, it writes
`UpToDateStatusType.UpToDateWithUpstreamTypes` if declaration output is unchanged.
The object already resides in the project-status cache. This is a build-status
retag, not a syntax-node retag or a fresh construction. It lets referenced
projects update output timestamps instead of rebuilding unchanged declarations.
Under the fixed-discriminant ruling this write needs an adaptation, such as
replacing the cached record with a newly constructed status. No adaptation is
made here.

Other outside entries update binding flags, aggregate node flags, merge symbols,
mark optional chains, update incremental signatures, recompute module indicators,
or temporarily replace and restore control-flow antecedents. Those purposes
are given with each row. No resolved write changes an existing syntax node's
`kind` to a different SyntaxKind.

## Locations: kind first

'''
md=intro+table([w for w in r['writes'] if w['property']=='kind'])+'\n\n## Other finite-literal discriminants\n\n'+table([w for w in r['writes'] if w['population']=='literal discriminant'])+'\n\n## Broader and partial checker candidates\n\n'+table([w for w in r['writes'] if w['population']=='partial checker discriminant'])+'''

## Unresolved dynamic-key writes and coverage limits

There are **201** bracket-write sites whose key type is not a finite union of
literal names. They are not included in the 437 resolved property events.
Many are array or dictionary slots. Some involve `any` or generic clones:
`cloneSourceFileWorker` and `cloneNode` use property-copy loops whose runtime
keys may include tag names. These loops skip properties already initialized on
the clone, but proving that skip for every reachable allocator and input needs
runtime-key/provenance reasoning. The stock checker cannot resolve a single
written property declaration there. This census does not call them zero writes.
Each site and its checked receiver/key type follows. Arbitrary user-defined
setters, aliases of reflective builtins, defineProperties descriptor maps,
prototype inheritance and mutations performed by called functions are not
expanded into implicit per-property writes. Consequently this is an exhaustive
ledger for the described AST/resolved-property rule, not a proof that every
possible runtime discriminant mutation has been counted.

| Write | Receiver type | Key type |
| --- | --- | --- |
'''
for w in r['unresolved']:md+='| '+' | '.join(map(esc,[place(w),w['receiver_type'],w['key_type']]))+' |\n'
md+='''
## Reproduction and validation

```
NODE_PATH=<stock-api-node_modules> node stage3/fixtures/assertions/discriminant-writes.cjs <prepared-TypeScript-root> <census-directory> <scratch-output> > <scan-log> 2>&1
python3 stage3/fixtures/assertions/discriminant-report.py <scratch-output>/discriminant-writes-raw.json stage3/fixtures/assertions > <report-log> 2>&1
python3 stage3/fixtures/assertions/discriminant-report.py <scratch-output>/discriminant-writes-raw.json <scratch-output>/mutant --mutant-omit-status-write > <mutant-log> 2>&1
```

Use the prepared upstream tree and stock npm environment from the Reproduction section of this bucket's
`README.md`. The scan finished with zero diagnostics and 437 resolved events,
201 unresolved dynamic-key sites. Report validation checks declaration presence,
unique write keys, the three exact constructor `kind` assignments, and the cached
status retag. The mutant removes only the status write from the real census input;
validation exits nonzero with `lost cached-status retag`. Three additional
mutants omit the Node constructor kind write, erase a resolved declaration,
and duplicate a write; each exits 1 at its corresponding invariant. This proves that an
omitted retag is caught by the report check. It does not prove escape analysis or
Adamic's new immutability rule, which this unit has not implemented or tested.
No fixtures, status.json, compiler code or another bucket were changed.
'''
(base/'DISCRIMINANT_WRITES.md').write_text(md)
print(json.dumps(r['summary'],indent=2));print('report invariants passed')
