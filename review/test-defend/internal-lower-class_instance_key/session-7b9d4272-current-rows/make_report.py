import pathlib,json,re,subprocess
p=pathlib.Path(__file__).parent
names=['TestClassWrongOutputKeysRepair','TestClassWrongOutputPrivateRepair','TestDefiniteAssignmentUsesReadiness']
subs=['TestClassFeaturesPrivateStorage','TestDefiniteAssignmentSoundNeighbors','TestReadinessElisionRequiresDominatingAssignment']
base=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip()
listed=[x for x in (p/'list.log').read_text().splitlines() if re.fullmatch('Test\\w+',x)]
old=[x for x in (p/'audit-list.log').read_text().splitlines() if re.fullmatch('Test\\w+',x)]
(p/'inventory.json').write_text(json.dumps({'base':base,'current':listed,'added_since_audit':sorted(set(listed)-set(old)),'vanished_since_audit':sorted(set(old)-set(listed))},indent=2))
def events(f):
 es=[]
 for line in f.read_text().splitlines():
  try: es.append(json.loads(line))
  except ValueError:pass
 return es
matrix={}
for ident in ['D1','D2','D3','D4','D5']:
 es=events(p/(ident+'.log'))
 if ident=='D5':
  for f in p.glob('D5-alone-*.log'):es+=events(f)
 status={}
 for e in es:
  if e['Action'] in ['pass','fail','skip'] and 'Test' in e and '/' not in e['Test']:status[e['Test']]=e['Action']
 out={n:'unknown' for n in listed};out.update(status)
 matrix[ident]={'rows':out,'rows_failed':sorted(n for n,s in out.items() if s=='fail'),'rows_passed':sorted(n for n,s in out.items() if s=='pass'),'rows_skipped':sorted(n for n,s in out.items() if s=='skip'),'unknown_rows':sorted(n for n,s in out.items() if s=='unknown'),'run':json.loads((p/(ident+'-run.json')).read_text())}
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
def failing(ident,name):
 files=[p/(ident+'.log')]
 if ident=='D5':files=[p/('D5-alone-'+name+'.log')]
 for f in files:
  for e in events(f):
   text=e.get('Output','').strip()
   if e.get('Test')==name and (re.search(r'\w+\.go:\d+:',text) or text.startswith('panic:')):return text
 return 'No failure; row passed.'
changes={'D1':('internal/lower/readiness.go:375','Change call detection result from true to false; calls no longer invalidate field readiness facts.'),'D2':('internal/lower/class.go:364','Change the external-receiver instantiation condition from OR to AND.'),'D3':('internal/lower/class_features.go:83','Change fresh-object classification to false, disabling the explicit-copy exemption.'),'D4':('internal/lower/class_static.go:210','Change the static-method receiver parameter from thisLocal(function) to -1.'),'D5':('internal/lower/class_static.go:177','Off-by-one the static dispatch table function index: function to function - 1.')}
# Lines are resolved against the starting source, rather than assumed from a scratch file.
needles={'D1':'calls = true','D2':') || lowered == nil {','D3':'fresh := argument.Kind == ast.KindObjectLiteralExpression','D4':'l.signature(function, member, l.thisLocal(function))','D5':'meta.Methods[slot] = function'}
for ident,(fileline,change) in list(changes.items()):
 file=fileline.split(':')[0];source=subprocess.check_output(['git','show',base+':'+file]).decode().splitlines();lines=[i+1 for i,s in enumerate(source) if needles[ident] in s];assert len(lines)==1
 changes[ident]=(file+':'+str(lines[0]),change)
rows=[]
for n,sub,ids,defense,unique in zip(names,subs,[['D3'],['D2','D4','D5'],['D1']],['defended','not defended','defended'],['D3',None,'D1']):
 last=unique or 'D4';command=matrix[last]['run']['command'];line=failing(last,n)
 rows.append({'test':n,'package':'internal/lower','prior_verdict':'subsumed','subsumed_by':[sub],'defense':defense,'unique_mutant':unique+' '+changes[unique][0] if unique else None,'attempts':[{'mutant':i,'file_line':changes[i][0],'change':changes[i][1],'rows_failed':matrix[i]['rows_failed']} for i in ids],'evidence':command+'; '+line})
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
report=f'''# Defense of class repairs and definite-assignment readiness

Starting origin/main: {base}. Audit base: 8171b3173bdbfce1f7982d3c4f731279307ece37.
Two requested rows defended; private repair not defended after three attempts.
No test, oracle or harness was changed. Production sources are restored.

## Code under test and oracle

Code under test: internal/lower Lower and its class storage, static/instance method dispatch, Object.keys classification, and readiness propagation. The class-repair oracles now compare source execution in Node against generated JavaScript execution in Node (stdout, exit, relevant stderr). These assertions have changed since the audit's absence-of-error descriptions. DefiniteAssignmentUsesReadiness uses handwritten counts of nonempty ir.Read/ir.Property readiness tags, plus Node agreement for its initialized controls. The count oracle does not identify which read carries a tag or check its text. None of the three rows is an executor twin or a cost row.

## Baseline and scope

npm ci --prefix stage3/api completed before baseline. Warm /workspace/adamic-tools/env.sh worked; setup skipped. nproc=5. Clean whole-package baseline passed in 38.695 test-binary seconds. All three assigned names and subsumers remain present. Current discovery has {len(listed)} tests, audit {len(old)}. Added and vanished names are in inventory.json. Every mutant's initial matrix used the full current package.

Baseline skips: TestOriginalCycleLedger needs pristine pinned upstream TypeScript and generated diagnostics; TestOptionalWideningCensus needs an external project's config and output path. Their bodies inventory import-cycle/type-relation rules rather than call Lower or the mutated functions. The interface_Node/readonly_Node[] subcase of TestMixedUnionContractGraph remains explicitly skipped awaiting support. No skipped case is asserted to have passed. D1 and D3 uniqueness is established among the completed available rows of the current package; no repository-wide claim is made.

## Coverage and semantic leads

For each requested row and its named subsumer, ran: timeout 120 go test -count=1 -timeout 90s -run '^NAME$' -coverpkg=./internal/lower -coverprofile=PATH ./internal/lower/ > PATH-coverage.log 2>&1. All six profiles passed. These instrument Go lowering only, not executed JavaScript or C. Exact exclusive coverage blocks are saved beside profiles.

KeysRepair has 262 covered blocks absent from PrivateStorage. Its semantic distinction is an explicit fresh string-key copy of a structurally viewed iterable, alongside a nominal iterable class. D3 disables the fresh-copy exemption; only KeysRepair fails, on the symbol-key-view refusal. PrivateRepair has 367 blocks absent from SoundNeighbors, including staticInstance and callOrMethod. It delegates a static method's instance parameter to an instance method reading private storage. D2 damages receiver instantiation but virtual dispatch still selects the right method in this input; the entire package passes. D4 removes the static receiver parameter and fails PrivateRepair plus PrivateAndPublicStaticsAgreeWithNode. D5 misindexes static dispatch and panics in both rows when run individually. DefiniteAssignmentUsesReadiness has 151 blocks absent from ReadinessElisionRequiresDominatingAssignment. Its last subcase writes a field, calls a function that rebinds the box, then reads the field. D1 suppresses call invalidation; only this row loses its required guard. The current named subsumer exercises eager TypeScript non-null assertions rather than this mutable field readiness history.

## Mutant observations

'''
for ident in changes:
 m=matrix[ident]; report+=f"{ident}: {changes[ident][0]}: {changes[ident][1]} Failed: {', '.join(m['rows_failed']) or 'none'}. Wall: {m['run']['wall_seconds']:.3f}s. Standalone diff: {ident}.diff; go vet log: {ident}-vet.log.\n\n"
report+='''All five standalone diffs passed go vet ./internal/lower/ and were generated from the starting source. D1, D2, D3 and D4 completed whole-package matrices. D5 aborted on a Go index-out-of-range panic, so its unfinished rows remain unknown in matrix.json. Reran all three requested rows and PrivateAndPublicStaticsAgreeWithNode alone under D5. Both class-repair neighbors and the readiness row were directly observed; the private row and the additional static-method row panic. No timeout occurred. Full passed lists for D1 and D3 are in matrix.json.

## Issues, limits, and owner finding

The requested 15 GB free-space threshold cannot be met on /tmp, whose total capacity is 8.8 GB. Initial /workspace free space was 4.0 GB. Deleted the earlier completed unit's scratch/cache directory under /tmp, never repository or tools. Rechecked disk and observed ample room for this unit's small artifacts; no disk failure occurred. The two mounts are distinct, so deleting /tmp cannot increase workspace capacity.

The audit's rows.json is a list of names; verdict objects live in results.json and REPORT.md. Assuming rows.json held objects caused two harmless parsing retries. Audit assertions are stale relative to this main: current class repairs execute Node agreement, and readiness gained a call-rebinding subcase. Default shell cwd is /workspace, so one read command needed its explicit repo workdir. A read-only network check failed through the restricted proxy; the authorized fetch succeeded with escalation. Existing defense branch contains a different pair of rows, so this session is nested and preserved alongside that prior evidence without force-pushing.

PrivateRepair is not defended by these three attempts: one survives and two are caught by another current row. Its name promises a working instance-delegation repair and its assertions check that generated JavaScript agrees with source Node on the printed secret. No name/assertion mismatch found. It does not exercise native execution, inheritance, borrowed receivers, or additional secrets; no deletion recommendation follows from this bounded attempt set. D2's survival is not claimed to expose unguarded output behavior. No oracle weakening, native-runtime mutation, additional package replay, or repository-wide uniqueness was attempted.
'''
(p/'REPORT.md').write_text(report)
print(json.dumps(rows,indent=2))
