import pathlib,json,subprocess,time,os,difflib
p=pathlib.Path('review/test-audit/internal-oracle-method_values_mutant')
base=json.load((p/'base.json').open()); plan=json.load((p/'plan.json').open()); names=json.load((p/'names.json').open())
def run(label,pattern,extra=None):
 env=os.environ.copy();env['ADAMIC_MUTANT']=label;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u063/cache/'+label
 if extra:env.update(extra)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]
 s=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 with (p/'runs.jsonl').open('a') as f:f.write(json.dumps({'id':label,'cmd':cmd,'seconds':time.monotonic()-s,'exit':r.returncode,'env':{k:env[k] for k in ['ADAMIC_MUTANT','ADAMIC_BUILD_CACHE_DIR']}})+'\n')
 return r.returncode
scoped='^('+'|'.join(names)+')$'
# Measure function coverage at the clean bounded baseline before edits.
with (p/'coverage.log').open('w') as f:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/'coverage.out'),'./internal/oracle/','-run',scoped],stdout=f,stderr=subprocess.STDOUT)
with (p/'coverage-functions.txt').open('w') as f:subprocess.run(['go','tool','cover','-func='+str(p/'coverage.out')],stdout=f,stderr=subprocess.STDOUT)
# Verify each standalone production diff in its original source, without a switch.
for m in plan:
 f=m['file'];path=pathlib.Path(f);path.write_text(base[f].replace(m['old'],m['new']))
 s=time.monotonic()
 with (p/(m['id']+'-vet.log')).open('w') as log:r=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
 with (p/'vet-times.jsonl').open('a') as log:log.write(json.dumps({'id':m['id'],'exit':r.returncode,'seconds':time.monotonic()-s})+'\n')
 path.write_text(base[f])
 assert r.returncode==0,m['id']
# Instrument only the already frozen changes.
reps={
'M01':('!l.provenModuleReads[node] && l.checked(local)','auditChoice("M01", !l.provenModuleReads[node] && l.checked(local), l.provenModuleReads[node] && l.checked(local))'),
'M02':('return true','return auditChoice("M02", true, false)'),
'M03':('Value: ir.BooleanConstant{Value: false}})','Value: ir.BooleanConstant{Value: auditChoice("M03", false, true)}})'),
'M04':('body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}})','body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: auditChoice("M04", true, false)}})'),
'M05':('l.result.Locals[local].NamespaceVar && declaration.Initializer() == nil','auditChoice("M05", l.result.Locals[local].NamespaceVar && declaration.Initializer() == nil, l.result.Locals[local].NamespaceVar && declaration.Initializer() != nil)'),
'M06':('"'+plan[5]['old']+'"','auditChoice("M06", "'+plan[5]['old']+'", "'+plan[5]['new']+'")'),
'M07':('Operator: ir.Equal, Left: ir.TypeOf{Value: held}','Operator: auditChoice("M07", ir.Equal, ir.NotEqual), Left: ir.TypeOf{Value: held}'),
'M08':('name = "number"','name = auditChoice("M08", "number", "string")'),
'M09':('known && held != narrowed && (held == ir.Object','known && auditChoice("M09", held != narrowed, held == narrowed) && (held == ir.Object'),
'M10':('b.read(b.parameters[len(checks)])','b.read(b.parameters[auditChoice("M10", len(checks), len(checks)-1)])')}
changed={f:t for f,t in base.items() if f.startswith('internal/lower/')}
for m in plan:
 old,new=reps[m['id']]; f=m['file'];assert changed[f].count(old)==1,(m['id'],old);changed[f]=changed[f].replace(old,new)
changed['internal/lower/lower.go']=changed['internal/lower/lower.go'].replace('func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n if auditSelector() == "P01" { return nil, nil }')
helper='package lower\nimport "os"\nfunc auditSelector() string {return os.Getenv("ADAMIC_MUTANT")}\nfunc auditChoice[T any](id string, original, changed T) T {if auditSelector()==id{return changed};return original}\n'
pathlib.Path('internal/lower/audit_mutant.go').write_text(helper)
for f,t in changed.items():pathlib.Path(f).write_text(t)
subprocess.run(['gofmt','-w',*changed,'internal/lower/audit_mutant.go'],check=True)
(p/'switched-source').mkdir(exist_ok=True)
for f in [*changed,'internal/lower/audit_mutant.go']:(p/'switched-source'/ (pathlib.Path(f).name+'.fixture')).write_text(pathlib.Path(f).read_text())
# Include the generic fixture family, limited to the inputs registering these code paths.
fixtures='library_method_values|module_namespace_reads|namespace-live-export|namespaces|union_narrow|narrowed_union'
pattern=scoped+'|^TestNativeAgreesWithNode/.*('+fixtures+')'
(p/'matrix-pattern.txt').write_text(pattern)
for m in plan:
 run(m['id'],pattern)
 events=[]
 for line in (p/(m['id']+'.log')).open():
  try:events.append(json.loads(line))
  except:pass
 if any(e.get('Output','').startswith('panic:') for e in events):
  for name in names:run(m['id']+'-'+name,'^'+name+'$',{'ADAMIC_MUTANT':m['id']})
# Only ordinary rows receive production-entry probes. Witnesses have no vacuity judgement.
for name in [names[4],names[7],names[12]]:run('P01-'+name,'^'+name+'$',{'ADAMIC_MUTANT':'P01'})
for f,t in changed.items():pathlib.Path(f).write_text(base[f])
pathlib.Path('internal/lower/audit_mutant.go').unlink()
# Authorized harness weakening for witnesses, with no compiler edits.
f='internal/oracle/oracle_test.go';original=base[f]
changed=original.replace('func disagreement(oracle run, native run) string {','func disagreement(oracle run, native run) string {\n if os.Getenv("ADAMIC_MUTANT")=="W01" {return ""}')
pathlib.Path(f).write_text(changed)
(p/'W01.switched.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
run('W01',scoped)
pathlib.Path(f).write_text(original)
f='internal/oracle/method_values_ownership_test.go';original=pathlib.Path(f).read_text()
changed=original.replace('strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free")','(false && strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free"))').replace('strings.Contains(string(report.stderr), "LeakSanitizer: detected memory leaks")','(false && strings.Contains(string(report.stderr), "LeakSanitizer: detected memory leaks"))')
pathlib.Path(f).write_text(changed)
(p/'diffs/W02.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
run('W02','^TestLibraryMethodCaptureMutants$')
pathlib.Path(f).write_text(original)
f='internal/oracle/namespaces_test.go';original=pathlib.Path(f).read_text();changed=original.replace('observed.exitCode == expected.exitCode','observed.exitCode == expected.exitCode || true')
pathlib.Path(f).write_text(changed)
(p/'diffs/W03.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
run('W03','^TestNamespaceStateMutants/skip_ready_check$')
pathlib.Path(f).write_text(original)
# Standalone weakened differential comparator, without selector.
changed=base['internal/oracle/oracle_test.go'].replace('switch {\n\tcase oracle.exitCode != native.exitCode:', 'switch {\n\tcase false && oracle.exitCode != native.exitCode:').replace('case !bytes.Equal(oracle.stdout, native.stdout):','case false && !bytes.Equal(oracle.stdout, native.stdout):').replace('case !bytes.Equal(oracle.stderr, native.stderr):','case false && !bytes.Equal(oracle.stderr, native.stderr):')
(p/'diffs/W01.diff').write_text(''.join(difflib.unified_diff(base['internal/oracle/oracle_test.go'].splitlines(True),changed.splitlines(True),fromfile='a/internal/oracle/oracle_test.go',tofile='b/internal/oracle/oracle_test.go')))
# Replay W01 standalone to prove exact diff, rather than only selector equivalence.
pathlib.Path('internal/oracle/oracle_test.go').write_text(changed)
run('W01-standalone',scoped)
pathlib.Path('internal/oracle/oracle_test.go').write_text(base['internal/oracle/oracle_test.go'])
for wid in ['W01','W02','W03']:
 subprocess.run(['git','apply',str(p/'diffs'/ (wid+'.diff'))],check=True)
 with (p/(wid+'-vet.log')).open('w') as f:r=subprocess.run(['go','vet','./internal/oracle/'],stdout=f,stderr=subprocess.STDOUT)
 subprocess.run(['git','apply','-R',str(p/'diffs'/ (wid+'.diff'))],check=True)
 assert r.returncode==0,wid
print('audit runs complete, production and harness restored')
