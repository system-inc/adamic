import subprocess,time,json,os,shutil
from pathlib import Path
out=Path(__file__).resolve().parent
root=out.parents[2]
rows=[]
def run(name,args,expected,needle,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  try:code=subprocess.run(args,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 text=(out/(name+'.log')).read_text(); good=code==expected and needle in text
 row=dict(name=name,command=args,exit=code,seconds=round(time.monotonic()-start,3),caught=good);rows.append(row)
 (out/'extra-mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 if not good:raise SystemExit(1)
def overlay(name,file,source):
 mutant=out/(name+'.go.txt');mutant.write_text(source)
 path=out/(name+'.json');path.write_text(json.dumps({'Replace':{str(root/file):str(mutant)}}))
 return str(path)
def test(pkg,name,ov=None):
 args=['go','test']
 if ov:args+=['-overlay',ov]
 return args+['./'+pkg,'-run',name,'-count=1','-timeout','90s','-v','-parallel','4']
file='internal/lower/records_named.go';source=(root/file).read_text();before='if l.includesNull(t) && !dynamicObjectType(t) {';assert before in source
file='internal/lower/records_divergence.go';source=(root/file).read_text();before='func (l *lowering) recordDivergenceRefusal(node *ast.Node) error {\n';assert before in source
ov=overlay('refusal-removal',file,source.replace(before,before+'\treturn nil // deliberate mutant\n',1));env=os.environ.copy();env['ADAMIC_RECORD_REFUSAL_MUTANT']='1'
run('refusal-removal',test('internal/oracle','^(TestRecordRefusal.*|TestRecordPrototypeMutant|TestRecordNarrowingMutant|TestNamedRecordTypeGuardMutant)$',ov),0,'refusal removal admits divergent checked stop',env)
file='internal/flow/flow_test.go';source=(root/file).read_text();before='"../oracle/testdata/*.a",';assert before in source
run('flow-refused-directory',test('internal/flow','^TestFlowCorpusRemainder$',overlay('flow-refused-directory',file,source.replace(before,before+'\n\t\t"../oracle/testdata/records_refused/*.a",',1))),1,'records_compare_properties_left')
env=os.environ.copy();env['ADAMIC_FAMILY_PROBE']='TestCountsAreRecorded=internal/oracle/testdata/records_operations.a'
run('counts-mutant',test('internal/oracle','^TestCountsAreRecorded_[0-9]+$'),1,'the counts changed',env)
fixtures=Path('/workspace/scratch/loom-r-records-ac-m-stage3-mutant');fixtures.mkdir(exist_ok=True);target=fixtures/'records';shutil.copytree(root/'stage3/fixtures/records',target,dirs_exist_ok=True)
status=target/'status.json';data=json.loads(status.read_text());[row.update(stage0={'outcome':'Compiles','what':''}) for row in data if row['file']=='07_optional_view.a'];status.write_text(json.dumps(data,indent=2)+'\n');(out/'stage3-mutant-status.json').write_text(status.read_text())
run('stage3-classification',test('stage3/fixtures','^TestFixturesRecords$')+['-args','-fixtures',str(fixtures)],1,'07_optional_view.a')
