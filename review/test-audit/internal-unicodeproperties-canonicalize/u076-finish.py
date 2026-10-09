import pathlib,json,subprocess,time,difflib,os
p=pathlib.Path('review/test-audit/internal-unicodeproperties-canonicalize');file=pathlib.Path('internal/unicodeproperties/canonicalize_test.go');before=file.read_text();sig='func unicodeNodeBatches(lines []string, batchSize, workers int, run func(string) (int, []string, error)) ([]unicodeNodeResult, error) {'
plan=[dict(id='S02',old='results[result.index] = result',new='results[(result.index+1)%len(results)] = result'),dict(id='S03',old=sig,new=sig+' if true {return nil,nil}')];results=[]
for q in plan:
 q['file']=str(file);q['line']=before[:before.index(q['old'])].count('\n')+1;after=before.replace(q['old'],q['new']);(p/'probes'/(q['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+str(file),tofile='b/'+str(file))))
 try:
  file.write_text(after)
  with (p/'logs'/('validate-'+q['id']+'.log')).open('w') as out:subprocess.run(['go','vet','./internal/unicodeproperties/'],stdout=out,stderr=subprocess.STDOUT,check=True)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run','^TestUnicodeNodeBatchOrderAndLimit$'];start=time.monotonic()
  with (p/'logs'/(q['id']+'.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  es=[json.loads(l) for l in (p/'logs'/(q['id']+'.log')).read_text().splitlines() if l.startswith('{')];q.update(command=' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,events=es);results.append(q)
 finally:file.write_text(before)
(p/'setup-probes.json').write_text(json.dumps(results,indent=2))
regex='^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-'
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run',regex]
with (p/'logs'/'restored.log').open('w') as out:subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,check=True)
