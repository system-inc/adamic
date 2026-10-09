from pathlib import Path
import json,subprocess,os,time,difflib
p=Path('review/test-defend/internal-lower-library_node_fs_file/session-7b9d4272')
plan=[('D04', 'internal/lower/invariance.go', 'source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)', 'source, target := l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)\n\t\tif viewed.Flags & ast.SymbolFlagsOptional != 0 && l.checker.TypeToString(source) == "boolean" { return nil }', 'Return early from mutable-property proof for optional views of boolean fields', 'widening')]
(p/'extra-plan.json').write_text(json.dumps(plan,indent=2)+'\n');results=json.loads((p/'results.json').read_text())
for mid,file,old,new,change,target in plan:
 orig=Path(file).read_text();assert orig.count(old)==1,(mid,orig.count(old))
 patch=p/(mid+'.diff');patch.write_text(''.join(difflib.unified_diff(orig.splitlines(True),orig.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 subprocess.run(['git','apply','--check',str(patch)],check=True);subprocess.run(['git','apply',str(patch)],check=True)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/fsfile-defend/cache/'+mid;start=time.monotonic()
 try:
  with open(p/(mid+'-vet.log'),'w') as log:vet=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  row={'mutant':mid,'target':target,'file_line':file+':'+str(orig[:orig.index(old)].count('\n')+1),'change':change,'vet_exit':vet}
  if vet==0:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];lp=p/(mid+'.log')
   with open(lp,'w') as log:rc=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
   es=[]
   for line in lp.read_text().splitlines():
    try:es.append(json.loads(line))
    except:pass
   row.update(exit=rc,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),failed=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test')],passed=[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test')],skipped=[e['Test'] for e in es if e.get('Action')=='skip' and e.get('Test')],cooked=any('test timed out' in e.get('Output','') for e in es),diagnostics=[{'test':e.get('Test'),'line':e['Output'].strip()} for e in es if e.get('Test','').startswith(('TestNodeFSFileDoesNotAuthorizeMutableWidening','TestNodeFSFileRefusesVoidValues','TestNodeFSFileRefusesPinnedUnsupportedOverloads')) and '.go:' in e.get('Output','')])
  row['wall_seconds']=round(time.monotonic()-start,3);results.append(row);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(mid,vet,row.get('failed'),flush=True)
 finally:subprocess.run(['git','apply','-R',str(patch)],check=True)
