import json,subprocess,time,os,shutil,difflib
from pathlib import Path
p=Path(__file__).resolve().parent
base=Path('stage1/cohere/estree')
original={f:f.read_text() for f in [base/'estree_test.go',base/'stalls_test.go',base/'stalls_deadlines_test.go',base/'syntax_mutants_self_setup_test.go']}
plan=[
 ('W01','estree_test.go','func firstDifference(want, got []byte) string {','func firstDifference(want, got []byte) string {\n if os.Getenv("ADAMIC_AUDIT") == "W01" { return "" }','func firstDifference(want, got []byte) string {\n return ""'),
 ('W02','stalls_test.go','func portStallControlResult(timedOut bool, err error) error {','func portStallControlResult(timedOut bool, err error) error {\n if os.Getenv("ADAMIC_AUDIT") == "W02" { return nil }','func portStallControlResult(timedOut bool, err error) error {\n return nil'),
 ('W03','syntax_mutants_self_setup_test.go','if !strings.Contains(string(got), "0 Program ") {','if os.Getenv("ADAMIC_AUDIT") != "W03" && !strings.Contains(string(got), "0 Program ") {','if false && !strings.Contains(string(got), "0 Program ") {'),
 ('S01','stalls_deadlines_test.go','Setpgid: true','Setpgid: os.Getenv("ADAMIC_AUDIT") != "S01"','Setpgid: false'),
 ('S02','syntax_mutants_self_setup_test.go','return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)','return os.WriteFile(filepath.Join(dir, auditArtifact("S02", "oracle")), data, 0755)','return os.WriteFile(filepath.Join(dir, "missing"), data, 0755)'),
 ('S03','syntax_mutants_self_setup_test.go','return os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(program)), 0644)','return os.WriteFile(filepath.Join(dir, auditArtifact("S03", "port.c")), []byte(native.C(program)), 0644)','return os.WriteFile(filepath.Join(dir, "missing"), []byte(native.C(program)), 0644)'),
 ('S04','syntax_mutants_self_setup_test.go','return native.Build(string(data), filepath.Join(dir, "port"), estreeFamilyNativeOptions())','return native.Build(string(data), filepath.Join(dir, auditArtifact("S04", "port")), estreeFamilyNativeOptions())','return native.Build(string(data), filepath.Join(dir, "missing"), estreeFamilyNativeOptions())'),
 ('S05','syntax_mutants_self_setup_test.go','return os.WriteFile(filepath.Join(dir, "ready.json"), data, 0644)','return os.WriteFile(filepath.Join(dir, auditArtifact("S05", "ready.json")), data, 0644)','return os.WriteFile(filepath.Join(dir, "missing"), data, 0644)')]
statuses=[]
try:
 for id,name,anchor,switch,plain in plan:
  f=base/name;s=f.read_text();assert s.count(anchor)==1;f.write_text(s.replace(anchor,switch,1))
  old=original[f];new=old.replace(anchor,plain,1)
  if id in ['W01','W02']:
   start=old.index(anchor);opening=start+len(anchor)-1;depth=1;end=opening+1
   while depth:
    if old[end]=='{':depth+=1
    elif old[end]=='}':depth-=1
    end+=1
   new=old[:opening+1]+'\n return '+('\"\"' if id=='W01' else 'nil')+'\n}'+old[end:]
   if id=='W02':new=new.replace('\n\t\"fmt\"','')
  (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
 f=base/'syntax_mutants_self_setup_test.go';s=f.read_text();s=s.replace('const testSyntaxMutantsShards = 3','func auditArtifact(id, name string) string { if os.Getenv("ADAMIC_AUDIT") == id { return "missing" }; return name }\n\nconst testSyntaxMutantsShards = 3');f.write_text(s)
 subprocess.run(['gofmt','-w']+[str(f) for f in original],check=True)
 cache=Path('/tmp/u088/cache/harness');env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(cache);env['ADAMIC_ESTREE_LIBRARY']='/tmp/u088/library'
 def run(id,regex,env):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex];start=time.monotonic()
  with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  statuses.append({'id':id,'regex':regex,'wall':time.monotonic()-start,'exit':r.returncode,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'command':cmd});(p/'harness-status.json').write_text(json.dumps(statuses,indent=2));print(id,statuses[-1],flush=True)
  return r.returncode
 for i in range(3):
  if run('harness-warm-'+str(i),'^TestProduct_SyntaxMutantsSetup_'+str(i).zfill(3)+'$',env):raise SystemExit('warm construction failed')
 for id,regex in [('W01','^(TestBoundedPortParserPlantedDisagreement|TestSyntaxMutants_[0-9]{3}|TestSyntaxMutantsUnion)$'),('W02','^(TestPortStallControlPlantedSurvivor|TestPortStallControl_[0-9]{3})$'),('W03','^TestSyntaxMutants_00[12]$'),('S01','^TestEstreeScopedProcessGroupDeadline$')]:
  current=env.copy();current['ADAMIC_AUDIT']=id;run(id,regex,current)
 for id,prefix,regex in [('S02','syntax-mutants-go-oracle','^TestProduct_SyntaxMutantsGoOracle$'),('S03','estree-syntax-mutant-emitted','^TestProduct_SyntaxMutantLowered_[0-9]{3}$'),('S04','estree-syntax-mutant-native','^TestProduct_SyntaxMutantNative_00[12]$'),('S05','syntax-mutant-ready','^TestProduct_SyntaxMutantsSetup_[0-9]{3}$')]:
  dest=Path('/tmp/u088/cache/'+id);shutil.copytree(cache,dest,copy_function=os.link,dirs_exist_ok=True)
  removed=[]
  for meta in dest.glob('*.inputs'):
   if meta.read_text().splitlines()[0].startswith('name '+prefix):
    target=dest/meta.stem
    if target.exists():shutil.rmtree(target)
    meta.unlink();removed.append(meta.stem)
  (p/(id+'-invalidated.json')).write_text(json.dumps(removed))
  current=env.copy();current['ADAMIC_AUDIT']=id;current['ADAMIC_BUILD_CACHE_DIR']=str(dest);run(id,regex,current)
finally:
 for file,data in original.items():file.write_text(data)
