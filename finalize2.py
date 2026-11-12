import pathlib,json,subprocess,hashlib,tarfile,re,shutil
out=pathlib.Path(__file__).parent;repo=pathlib.Path('/workspace/adamic')
def read(name):return json.loads((out/name).read_text())
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
control=read('control2-lane-summary.json');after=read('after2-lane-summary.json');reference=read('helper2-reference-census.json');public=read('public2-contracts.json');trace=read('repeated-trace2.json');identity=read('after3-identity.json');host=read('host2-declarations.json');mutant=read('restoring2-mutant.json')
assert control['allPass'] and after['allPass'] and control['shards']==after['shards']==398
assert control['memoryEventsUnchanged'] and after['memoryEventsUnchanged'] and control['oldCatchExecutions']==0
assert read('catch2-positive.json')['catchExecutions']==1
assert reference['requiredClosureConditionHolds'] and reference['summary']['before']==dict(references=9,catchReferences=3,otherReferences=6) and reference['summary']['after']==dict(references=6,catchReferences=0,otherReferences=6)
for name in reference['allTenFiles']:
 a=next(r for r in reference['results'] if r['label']=='before' and r['file']==name);b=next(r for r in reference['results'] if r['label']=='after' and r['file']==name)
 assert [(s['sourceLine'],s['referenceExpression']) for s in a['sites'] if not s['inStartTracingCatch']]==[(s['sourceLine'],s['referenceExpression']) for s in b['sites'] if not s['inStartTracingCatch']]
assert public['runtimeExportCount']==2256 and public['runtimeExportKeysIdentical'] and public['errorMessageIsRuntimeExport']
assert trace['controlSelfInventoryDiffers'] and trace['deterministicInventoryEqual'] and trace['typesAndLegendByteIdentical']
assert identity['officialReplayByteIdentical'] and identity['sourceIdenticalToAfter2'] and identity['emittedByteIdenticalToAfter2']
assert host['onlyImportsAndFunctionDeclarations'] and host['beforeAfterByteIdentical'] and mutant['firstStopRestored']=='tracing.ts:381:1'
assert all(r['exit']==0 for r in read('root2-commands.json')) and all(r['exit']==0 for r in read('between-oracles2-commands.json')) and all(r['exit']==0 for r in read('guard-fixed-binding-commands.json'))
head=git('rev-parse','HEAD');parent=git('rev-parse','HEAD^');assert parent.startswith('2d7a1683') and git('rev-list','--count',parent+'..HEAD')=='1'
changed=git('diff','--name-only',parent,'HEAD').splitlines();assert all(n.startswith(('stage3/adapt/47-host-errors/','stage3/adapt/65-temporary-node-builtins/')) or n=='stage3/upstream/LEDGER.md' for n in changed)
base_ledger=subprocess.check_output(['git','show',parent+':stage3/upstream/LEDGER.md'],cwd=repo);ledger=(repo/'stage3/upstream/LEDGER.md').read_bytes();assert ledger.startswith(base_ledger);assert b'99-correctness-tracing-length' in base_ledger
old=pathlib.Path('/workspace/scratch/tracing-control2');new=pathlib.Path('/workspace/scratch/tracing-after3');files=sorted(p.relative_to(old) for p in (old/'src').rglob('*') if p.is_file());source_changes=[str(p) for p in files if (old/p).read_bytes()!=(new/p).read_bytes()];assert source_changes==['src/compiler/tracing.ts']
pid='declare const process: { readonly pid: number };';assert (old/'src/compiler/tracing.ts').read_text().count(pid)==(new/'src/compiler/tracing.ts').read_text().count(pid)==1
census=[]
for label in ['control2','after2']:
 data=[json.loads(s) for s in (out/(label+'-full-census.jsonl')).read_text().splitlines()];assert not data[0]['diagnostics'] and data[-1]['status']=='relation_report';census.append(dict(tree=label,checkerDiagnostics=0,completeTrailer=True))
stops=[json.loads(s) for s in (out/'after2-core-first-stop.jsonl').read_text().splitlines()];assert all(s['kind']=='NotYet' and s['where'].endswith('/sys.ts:1654:13') and s['reason']=='node:inspector.Session.post' for s in stops)
archives=[]
for label in ['control','after']:
 root=pathlib.Path('/workspace/scratch/tracing2-lane-'+label);target=out/(label+'2-lane.tgz')
 with tarfile.open(target,'w:gz',compresslevel=6) as archive:archive.add(root,arcname=root.name)
 archives.append(dict(file=target.name,bytes=target.stat().st_size,sha256=hashlib.sha256(target.read_bytes()).hexdigest()))
summary=dict(commit=head,parent=parent,oneCommit=True,changedFiles=changed,adaptedSourceFiles=len(files),adaptedSourceChanges=source_changes,catchReferencesBefore=3,catchReferencesAfter=0,exportReferencesBefore=6,exportReferencesAfter=6,exportReferenceBytesIdentical=True,runtimeExportKeys=2256,control=control,after=after,trace=trace,public=public,identity=identity,censuses=census,newCoreFirstStop='sys.ts:1654:13, node:inspector.Session.post',restoringMutant='tracing.ts:381:1 on 2d7a1683; this maps to tracing.ts:380:1 on the previous base',ledgerAppendedAfterBothExistingEntries=True,archives=archives)
(out/'FINAL_REPORT.json').write_text(json.dumps(summary,indent=2)+'\n')
if (out/'README.md').exists():shutil.copyfile(out/'README.md',out/'closure-stop.md')
(out/'README.md').write_text(f'''# tracing-fs-import completed proof

Task nqp4p7g. Final commit `{head}` is one commit on `{parent}`
(`typescript/tracing-fs-views-2` at 2d7a1683). It changes only adaptations 47,
65, and the appended ledger entry, after both the pop/popAll and fs-view entries.

## Corrected reference contract

| Quantity | Before | After |
| --- | ---: | ---: |
| Reads in startTracing catches | 3 | 0 |
| Live export reads in run.js and typescript.js | 6 | 6 |
| Runtime export keys, including errorMessage | 2,256 | 2,256 |

All six export expressions and their complete emitted source lines are byte
identical. `_tsc.js` alone eliminates errorMessage: its sole reader was the
removed catch. See helper2-reference-census.json and public2-contracts.json.

## Emission and module effects

All ten emitted files are inventoried. emit2-hunk-classification.md lists all
68 hunks, backed by the passing symbol/token certificate. Directly reviewed
tracing import/local/loader edits are marked. Outside changes are consistent
esbuild renames, the closed helper removal, or declaration-only hostErrors
relocation. Binding shorthand normalization preserves each property key and
checks the renamed local's symbol bijectively. The empty esbuild hostErrors
initializer callback contains only a strict directive, and its binding is
used only by no-argument calls.

host2-declarations.json pins the identical source module body: function
declarations errorCode, errorMessage, propertyOf, with no other top-level
statements. 47 drops the unused errorMessage import. All 17 fs reads bind to
the namespace import. Node 24.19.0 on Linux x64 has fs built in; normal builtin
loading cannot fail in these configurations. The source module graph and
actual load stacks in catch2-positive.json show sys's fs load before tracing
can be invoked. The synthetic loader-failure counter control fires once.

## Complete serial lanes

Both unmodified measure_shards.py invocations execute all 398 units with
concurrency 2, with no other build/test workload alongside the accepted pair.
Every shard passes. Each run records 106,366 passing observations, the same
single sanctioned Public APIs failure, and zero pending. All shard runner
copies match the input hash. Memory-event counters remain unchanged.
Control: {control['wallSeconds']:.3f}s, maximum shard {control['maximumShardSeconds']:.3f}s.
Final: {after['wallSeconds']:.3f}s, maximum shard {after['maximumShardSeconds']:.3f}s.
The control counter observes 398 harness processes and zero catch executions.
Raw reports, observations, logs and baseline diffs are in control2-lane.tgz
and after2-lane.tgz; archive hashes are in FINAL_REPORT.json. Each completed
private adapted-tree scratch view was removed only after its child exited and
its report was complete; its runner hash was recorded before removal.

## Trace and plain Node witnesses

Two control and two final compilations use the same strict 152-line project.
The controls themselves differ: trace event counts are 30 and 31. Final counts
are 31 and 33. tracing.ts samples ph X complete events when [time,endTime)
crosses a 10 ms sampling point. Removing only that sampled kind leaves the
same 23-event inventory in every run after named timestamp/pid normalization.
Observed sampled names are checkExpression, checkVariableDeclaration,
emitDeclarationFileOrBundle, emitJsFileOrBundle, findSourceFile,
processRootFiles, transformNodes. types.json and legend.json are byte-identical
across all four runs. Raw files and hashes are retained.

Ordinary project-mode tsc does not flush legend.json. The trace-only Node
wrapper evaluates the unchanged built _tsc.js bytes and intercepts its exit
to call the existing dumpTracingLegend before returning the compiler's actual
exit code. It changes no compiler source or clock. The separate plain witness
runs the real unmodified tsc.js normally: exits, stdout, stderr, emitted JS and
declarations all match. Both public d.ts files match byte for byte.

## Application, replay, census and mutant

Fresh official control/final apply, builds, harnesses and all 79-root adamic
types checks pass. Official final replay reproduces all 747 source files;
immediate 47 and 65 completed-state replays perform zero edits. Whole-owner
drift, module collision, parameter capture and bare namespace value controls
refuse. FINAL_REPORT.json links the current source/emit identity checks.

Both complete tracing censuses have zero checker diagnostics and complete
trailers. The require refusal disappears; the readonly process.pid declaration
is unchanged. Core's new first stop is sys.ts:1654:13,
node:inspector.Session.post. Restoring the exact control tracing.ts bytes brings
back the ambient-require refusal. On the new base its physical address is
381:1, corresponding to 380:1 before the pop amendment added a line.

## Retained failed attempts

Initial and superseded stop reports remain in initial-stop.md, closure-stop.md
and Git history. The old all-before-in-catch / after-zero condition was corrected
by the user; no live exports are deleted. Earlier strict raw trace comparison
and the first missing-legend wrapper attempts are retained without passing
claims. A full-project shadow probe was killed while an assertion formatted
cyclic compiler symbols; the guard now compares identity as a boolean and all
four direct binding controls pass with small messages. The overlapping first
control lane completed 398 passing shards but was rejected by the unchanged
resource guard because oom_kill increased. That full attempt is preserved in
control2-lane-attempt1.tgz. The accepted serial pair was rerun only after the
fix and all other checks completed. No lane sanction, expectation, timeout,
fixture, manifest or memory guard was weakened.
''')
for p in out.rglob('*.go'):raise AssertionError('evidence Go file must end in .go.txt: '+str(p))
print('final evidence complete for',head)
