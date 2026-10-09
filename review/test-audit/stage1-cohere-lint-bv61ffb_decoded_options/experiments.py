import subprocess,json,time,difflib,os,signal
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-lint-bv61ffb_decoded_options'); base=(p/'base.txt').read_text().strip().split()[0]
menu=[
('M1','stage1/cohere/lint/settings.ts','toLowerCase()','toUpperCase()','production'),
('M2','stage1/cohere/lint/settings.ts','return values[0] ?? fallback;','return values[1] ?? fallback;','production'),
('M3','stage1/cohere/lint/main.ts','    linter.run();','', 'production'),
('M4','internal/native/emit_branches.go','condition, whenTrue, whenNot)','condition, whenNot, whenTrue)','production'),
('P1','stage1/cohere/lint/main.ts','function run(row: string, countOnly: boolean): number {','function run(row: string, countOnly: boolean): number {\n    return 0;','probe'),
('P2','internal/native/emit.go','return cProgram(program, -1)','return ""','probe'),
('W1','stage1/cohere/lint/lint_test.go','len(commandDiagnostics(name, stderr.Bytes())) != 0','len(commandDiagnostics(name, stderr.Bytes())) < 0','witness'),
('W2','stage1/cohere/lint/bv61ffb_decoded_options_test.go','func decodedOptionsCheck(','func decodedOptionsCheck(','witness'),
('W3','stage1/cohere/lint/compiler_stage1_split_test.go','if diff := difference(got, want); diff != "" {','if diff := difference(got, want); diff == "u106-never" {','witness'),
('S1','stage1/cohere/lint/lint_test.go','if name == "go" {','if name != "go" {','setup'),
('S2','stage1/cohere/lint/compiler_stage1_split_test.go','const compilerAgreementSides = 5','const compilerAgreementSides = 6','setup'),
('S5','stage1/cohere/lint/compiler_stage1_split_test.go','if prepared.path == "" {','if prepared.path != "" {','setup'),
('S4','stage1/cohere/lint/bv61ffb_decoded_options_test.go','filepath.Join(output, "lint.c")','filepath.Join(output, "missing/lint.c")','setup'),
('S3','stage1/cohere/lint/complete_suggestion_isolated_shard_test.go','ADAMIC_COMPLETE_SUGGESTION_PLANT=0','ADAMIC_COMPLETE_SUGGESTION_PLANT=1','setup')]
production='^(TestDecodedOptionsAndMutant_[0-9]+|TestDecodedOptionsAndMutantUnion|TestClosedComparatorGaps|TestCompilerAndStage1Agree_(002|007|012))$'
patterns={'S5':'^TestProduct_CompilerAgreementGoOracle$','S4':'^TestDecodedOptionsAndMutant_Setup$','W1':'^TestExecuteFailsOnStderrOtherThanModuleDownloads$','W2':'^TestDecodedOptionsAndMutantPlantedFailure$','W3':'^TestCompilerAndStage1AgreePlantedDisagreement$','S1':'^TestCommandDiagnosticsDropOnlyModuleDownloads$','S2':'^TestCompilerAndStage1Agree_Setup$','S3':'^TestCompleteSuggestionSerialization_IsolatedShard$','P2':'^TestClosedComparatorGaps$'}
for mid,file,old,new,kind in menu:
 original=subprocess.check_output(['git','show',base+':'+file],text=True)
 if mid=='W2':
  import re
  m=re.search(r'func decodedOptionsCheck\([^\n]+\{',original);old=m.group();new=old+'\n\tif len(want) >= 0 { return nil }'
 assert old in original,(mid,old)
 changed=original.replace(old,new,1)
 (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 with (p/'menu.jsonl').open('a') as f:f.write(json.dumps({'id':mid,'file':file,'line':original[:original.index(old)].count('\n')+1,'old':old,'new':new,'kind':kind})+'\n')
 if os.getenv('PREPARE_ONLY'):continue
 Path(file).write_text(changed)
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u106/cache/'+mid
  if file.endswith('.go'):
   with (p/(mid+'-vet.log')).open('w') as out:
    rc=subprocess.call(['timeout','90','go','vet','./internal/native/' if file.startswith('internal/') else './stage1/cohere/lint/'],stdout=out,stderr=subprocess.STDOUT,env=env)
   if rc:raise RuntimeError('vet failed '+mid)
  cmd=['timeout','100','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',patterns.get(mid,production)]
  start=time.monotonic()
  with (p/(mid+'.log')).open('w') as out:
   proc=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True);rc=proc.wait()
   try:os.killpg(proc.pid,signal.SIGKILL)
   except ProcessLookupError:pass
  records=[]
  for line in (p/(mid+'.log')).read_text().splitlines():
   try:records.append(json.loads(line))
   except:pass
  result={'id':mid,'kind':kind,'command':cmd,'returncode':rc,'wall':time.monotonic()-start,'failures':[r['Test'] for r in records if r.get('Action')=='fail' and r.get('Test')],'completed':[r['Test'] for r in records if r.get('Action') in ('pass','fail','skip') and r.get('Test')],'output':[r.get('Output','') for r in records if 'Output' in r and ('Error' in r['Output'] or 'got ' in r['Output'] or 'survived' in r['Output'] or 'caught' in r['Output'] or 'want ' in r['Output'] or 'FAIL' in r['Output'])]}
  with (p/'results.jsonl').open('a') as out:out.write(json.dumps(result)+'\n')
 finally:Path(file).write_text(original)
