import pathlib,json,subprocess,time,os,signal,difflib
p=pathlib.Path('review/test-audit/internal-unicodeproperties-node');base=json.load((p/'base.json').open());plan=json.load((p/'plan.json').open())
scoped=json.load((p/'names.json').open())
rows=['TestVersionMatchesNode','TestNodeAgrees','TestNodeStringProperties','TestBinaryAliases','TestGeneralCategoryAliases','TestScriptAliasSet','TestRejectedNames','TestStringPropertyCensus','TestSetBoundaries','TestKnownMembership']
pattern='^('+'|'.join(rows)+')$';(p/'matrix-pattern.txt').write_text(pattern);json.dump(rows,(p/'matrix-rows.json').open('w'))
def run(label,pattern,selector=None):
 env=os.environ.copy();env['ADAMIC_MUTANT']=label if selector is None else selector
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run',pattern]
 s=time.monotonic()
 with (p/(label+'.log')).open('w') as f:
  child=subprocess.Popen(cmd,env=env,stdout=f,stderr=subprocess.STDOUT,start_new_session=True);code=child.wait()
  if code!=0:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
 with (p/'runs.jsonl').open('a') as f:f.write(json.dumps({'id':label,'seconds':time.monotonic()-s,'exit':code,'cmd':cmd,'selector':env['ADAMIC_MUTANT']})+'\n')
 return code
assert run('matrix-clean',pattern,selector='')==0
# Measure the likely subsumer's cost independently three times.
for i in range(3):assert run('time-TestStringPropertyCensus-'+str(i),'^TestStringPropertyCensus$',selector='')==0
# Standalone vet, each original file with only its one change.
for m in plan:
 f=m['file'];path=pathlib.Path(f);path.write_text(base[f].replace(m['old'],m['new']))
 s=time.monotonic()
 with (p/(m['id']+'-vet.log')).open('w') as log:r=subprocess.run(['go','vet','./internal/unicodeproperties/'],stdout=log,stderr=subprocess.STDOUT)
 with (p/'vet-times.jsonl').open('a') as log:log.write(json.dumps({'id':m['id'],'seconds':time.monotonic()-s,'exit':r.returncode})+'\n')
 path.write_text(base[f]);assert r.returncode==0,m['id']
f='internal/unicodeproperties/unicodeproperties.go';t=base[f]
t=t.replace('const (\n\tVersion','var (\n\tVersion')
t=t.replace('Version            = "17.0.0"','Version            = auditChoice("M02", "17.0.0", "17.0.1")')
t=t.replace('NodeUnicodeVersion = "17.0"','NodeUnicodeVersion = auditChoice("M01", "17.0", "16.0")')
t=t.replace('table = scriptExtensionNames','table = auditChoice("M03", scriptExtensionNames, scriptNames)')
t=t.replace('codePoint(generalCategoryNames, expression)','codePoint(auditChoice("M04", generalCategoryNames, binaryNames), expression)')
t=t.replace('if unicodeSets {','if auditChoice("M05", unicodeSets, !unicodeSets) {')
t=t.replace('Kind: KindCodePoints, Name: e.name','Kind: auditChoice("M06", KindCodePoints, KindStrings), Name: e.name')
t=t.replace('if equals > 1 || i == 0','if auditChoice("M08", equals > 1, equals > 0) || i == 0')
t=t.replace('expression[i+1:], true','expression[auditChoice("M09", i+1, i):], true')
t=t.replace('if ranges[mid].End < cp {','if auditChoice("M10", ranges[mid].End < cp, ranges[mid].End <= cp) {')
t=t.replace('ranges[lo].Start <= cp','auditChoice("M11", ranges[lo].Start <= cp, ranges[lo].Start < cp)')
t=t.replace('Sequences: s.sequences}','Sequences: auditChoice("M12", s.sequences, s.sequences[:len(s.sequences)-1])}')
t=t.replace('func Lookup(expression string, unicodeSets bool) (Property, bool) {','func Lookup(expression string, unicodeSets bool) (Property, bool) {\n if auditMutant=="P01" {return Property{}, false}')
pathlib.Path(f).write_text(t)
f='internal/unicodeproperties/tables.go';t=base[f].replace(plan[6]['old'],plan[6]['old'].replace('0x007F','auditChoice("M07", uint32(0x007F), uint32(0x007E))')).replace(plan[12]['old'],'var sequences_Basic_Emoji = []string{\n auditChoice("M13", "©️", "A"),')
pathlib.Path(f).write_text(t)
helper=pathlib.Path('internal/unicodeproperties/audit_mutant.go');helper.write_text('package unicodeproperties\nimport "os"\nvar auditMutant=os.Getenv("ADAMIC_MUTANT")\nfunc auditChoice[T any](id string, original, changed T) T {if auditMutant==id{return changed};return original}\n')
subprocess.run(['gofmt','-w','internal/unicodeproperties/unicodeproperties.go','internal/unicodeproperties/tables.go',str(helper)],check=True)
(p/'switched-source').mkdir(exist_ok=True)
for f in ['unicodeproperties.go','tables.go','audit_mutant.go']:(p/'switched-source'/(f+'.fixture')).write_text(pathlib.Path('internal/unicodeproperties/'+f).read_text())
try:
 for m in plan:
  mid=m['id'];run(mid,pattern)
  text=(p/(mid+'.log')).read_text()
  if 'panic: test timed out' in text:
   # The only serial expensive checker is already individually cooked; do not repeat its identical 90s scan.
   for row in rows:
    if row!='TestNodeAgrees':run(mid+'-'+row,'^'+row+'$',selector=mid)
  elif 'panic:' in text:
   for row in rows:run(mid+'-'+row,'^'+row+'$',selector=mid)
 run('P01',pattern)
finally:
 for f in ['internal/unicodeproperties/unicodeproperties.go','internal/unicodeproperties/tables.go']:pathlib.Path(f).write_text(base[f])
 if helper.exists():helper.unlink()
# Setup construction and witness weakenings are standalone authorized test-harness changes.
for special in json.load((p/'special-plan.json').open()):
 f=special['file'];old=base[f];changed=old.replace(special['old'],special['new']);assert old.count(special['old'])==1
 (p/'diffs'/(special['id']+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 pathlib.Path(f).write_text(changed)
 try:
  with (p/(special['id']+'-vet.log')).open('w') as log:r=subprocess.run(['go','vet','./internal/unicodeproperties/'],stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode==0
  pat='^TestUnicodeNodeShardCoverage$' if special['id']=='S01' else '^TestUnicodeNode(Shard|Stride)PlantedFailure$'
  run(special['id'],pat)
 finally:pathlib.Path(f).write_text(old)
# Probe standalone keeps its selector; ordinary production diffs do not.
f='internal/unicodeproperties/unicodeproperties.go';old=base[f];changed=old.replace('import "sort"','import ("sort"; "os")').replace('func Lookup(expression string, unicodeSets bool) (Property, bool) {','func Lookup(expression string, unicodeSets bool) (Property, bool) {\n if os.Getenv("ADAMIC_MUTANT")=="P01" {return Property{}, false}')
(p/'diffs/P01.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
pathlib.Path(f).write_text(changed)
try:
 with (p/'P01-vet.log').open('w') as log:r=subprocess.run(['go','vet','./internal/unicodeproperties/'],stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0
finally:pathlib.Path(f).write_text(old)
for diff in (p/'diffs').glob('*.diff'):subprocess.run(['git','apply','--check',str(diff)],check=True)
print('matrix and special checks complete; all sources restored')
