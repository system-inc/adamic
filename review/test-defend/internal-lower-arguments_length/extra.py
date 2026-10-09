import pathlib,subprocess,json,os,time,difflib
p=pathlib.Path('review/test-defend/internal-lower-arguments_length');menus=json.loads((p/'menu.json').read_text());runs=json.loads((p/'runs.json').read_text());names=json.loads((p/'matrix-selection.json').read_text())['tests'];extra_names=['TestProvenRelationsRefuse','TestWhatZeroOneRefusesIsRefusedWithAFix','TestUnknownReflectionRefusals'];scope=(p/'scope.log').read_text().splitlines();assert all(n in scope for n in extra_names);expanded=list(dict.fromkeys(names+extra_names));(p/'expanded-matrix-rows.json').write_text(json.dumps(expanded,indent=2)+'\n');extra=[]
orig={f:pathlib.Path(f).read_text() for f in ['internal/lower/library_array_predicate.go','internal/lower/census_small.go','internal/lower/invariance.go','internal/lower/unknown.go']}
def add(id,f,old,new,row,lead):
 assert old in orig[f],id;extra.append(dict(id=id,file=f,line=orig[f][:orig[f].index(old)].count('\n')+1,old=old,new=new,rows=[row],lead=lead,kind='change constant' if id=='D16' else 'return early'))
L='internal/lower/library_array_predicate.go';old='''if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.arrayPredicateDomain(member) {
				return false
			}
		}
		return true''';add('D16',L,old,old.rsplit('return true',1)[0]+'return false','TestArrayPredicatePreservesDeclaredElementContract','Exclusive union-domain success versus unknown-only reflection')
C='internal/lower/census_small.go';old=orig[C][orig[C].index('func (l *lowering) censusCallableParameterType('):];old=old[:old.index('\n}\n')+3];signature=old[:old.index('{')+1];add('D18',C,old,signature+'\n\treturn l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(parameter))\n}\n','TestOptionalFunctionValueRelation','Erase optional parameter undefined while leaving ordinary Animal/Dog relation unchanged')
(p/'extra-menu.json').write_text(json.dumps(extra,indent=2)+'\n')
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^('+'|'.join(extra_names)+')$']
with (p/'extra-baseline.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
assert r.returncode==0,'extra baseline red'
trials=[{**next(m for m in menus if m['id']==id),'id':id+'-replay','selection':extra_names} for id in ['D13','D15']]+[{**m,'selection':expanded} for m in extra]
try:
 for m in trials:
  id=m['id'];f=m['file'];s=orig[f];new=s.replace(m['old'],m['new'],1);diff=''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f));(p/(id+'.diff')).write_text(diff);subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],check=True);pathlib.Path(f).write_text(new)
  with (p/(id+'-vet.log')).open('w') as log:v=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  assert v.returncode==0
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^('+'|'.join(m['selection'])+')$'];env={**os.environ,'ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-arguments/cache/'+id};start=time.monotonic()
  with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  pathlib.Path(f).write_text(s)
  es=[json.loads(l) for l in (p/(id+'.log')).read_text().splitlines() if l.startswith('{')];failed=sorted({e['Test'].split('/')[0] for e in es if e.get('Test') and e['Action']=='fail'});passed=sorted({e['Test'] for e in es if e.get('Test') and '/' not in e['Test'] and e['Action']=='pass'});runs.append(dict(id=id,command=cmd,environment={'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},exit=r.returncode,seconds=time.monotonic()-start,rows_failed=failed,rows_passed=passed,errors=[e for e in es if e.get('OutputType')=='error']));(p/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');print(id,failed,len(passed),flush=True)
finally:
 for f,s in orig.items():pathlib.Path(f).write_text(s)
