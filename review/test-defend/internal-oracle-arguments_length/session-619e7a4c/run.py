from pathlib import Path
import subprocess,json,os,time,difflib
p=Path('review/test-defend/internal-oracle-arguments_length');selectors=json.loads((p/'selectors.json').read_text())
plan=[('D01','internal/native/runtime/adamic.c','write_all(adamic_stdout, bytes, length)','write_all(adamic_stdout, bytes, length - 1)','Off by one: large direct stdout writes omit their final byte','flush'),('D02','internal/native/runtime/adamic.c','// Everything the program printed comes first, as on Node.\n\tflush();','// Everything the program printed comes first, as on Node.','Drop the panic stdout flush statement','flush'),('D03','internal/native/runtime/adamic.c','write_all(adamic_stdout, bytes, length)','write_all(adamic_stderr, bytes, length)','Change large-write descriptor from stdout to stderr','flush'),('D04','internal/lower/cast_proof.go','for _, previous := range literals {','for _, previous := range literals[:max(0, len(literals)-1)] {','Off by one: omit the last previous tag from duplicate comparison','admission'),('D05','internal/native/emit_statements.go','if e.program.CallMayThrow(call) {','if !e.program.CallMayThrow(call) {','Flip direct void-call exception propagation condition','call'),('D06','internal/native/exceptions.go','e.nested(statement.Finally, nil)','/* finalizer statement omitted */','Drop the three explicit try-finally body emission statements','call'),('D07','internal/native/emit_functions.go','if expression.Direct > 0 {','if false {','Drop the direct closure-call fast path by flipping its condition','call')]
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n');rs=[];env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
for mid,f,old,new,change,target in plan:
 orig=Path(f).read_text();assert orig.count(old)==(3 if mid=='D06' else 1),(mid,orig.count(old));mod=orig.replace(old,new);dp=p/(mid+'.diff');dp.write_text(''.join(difflib.unified_diff(orig.splitlines(True),mod.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));subprocess.run(['git','apply','--check',str(dp)],check=True);subprocess.run(['git','apply',str(dp)],check=True);env['ADAMIC_BUILD_CACHE_DIR']='/tmp/arguments-defend/cache/'+mid;start=time.monotonic()
 try:
  pkg='internal/lower' if '/lower/' in f else 'internal/native'
  with open(p/(mid+'-vet.log'),'w') as log:vet=subprocess.run(['timeout','90','go','vet','./'+pkg+'/'],stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  row={'mutant':mid,'file_line':f+':'+str(orig[:orig.index(old)].count('\n')+1),'change':change,'target':target,'vet_exit':vet,'runs':[]}
  if vet==0:
   for i,s in enumerate(selectors):
    cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',s];lp=p/(mid+f'-{i}.log')
    with open(lp,'w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
    es=[]
    for line in lp.read_text().splitlines():
     try:es.append(json.loads(line))
     except:pass
    row['runs'].append({'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'exit':r.returncode,'failed':[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test')],'passed':[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test')],'skipped':[e['Test'] for e in es if e.get('Action')=='skip' and e.get('Test')],'cooked':any('test timed out' in e.get('Output','') for e in es),'diagnostics':[e for e in es if e.get('Action')=='output' and '.go:' in e.get('Output','') and e.get('Test','').startswith(('TestCallTargetThrowAgreesWithNode','TestDirectClosureCallAgreesWithNode','TestCheckedCastFlushesOutput','TestUncheckableCastAdmission'))]})
  row['wall_seconds']=round(time.monotonic()-start,3);rs.append(row);(p/'results.json').write_text(json.dumps(rs,indent=2)+'\n');print(mid,vet,[r['failed'] for r in row['runs']],flush=True)
 finally:subprocess.run(['git','apply','-R',str(dp)],check=True)
