from pathlib import Path
import subprocess,os
cases=[('record-discard-release','internal/native/view_callables_discard.go',' || %s->result == %d) adamic_release',' ) adamic_release','^TestCheckedViewStoredMarkerCalls$/record$'),('record-nullable-mask','internal/lower/view_nullish.go','of = ir.Object','of = ir.Record','^TestCheckedViewDictionaryContainerBatch$/wildcard-directories-good$'),('record-nullable-storage','internal/native/runtime/view_nullish.c','||actual==14','||false','^TestCheckedViewDictionaryArrays$/paths-producer-good$')]
for name,file,before,after,pattern in cases:
 p=Path(file);original=p.read_text();assert original.count(before)==1
 mutant=original.replace(before,after)
 if name=='record-discard-release':mutant=mutant.replace(', recorded, ir.Record, result)',', result)')
 try:
  p.write_text(mutant)
  logfile=Path('/tmp/views-integration-dictionary-deec-mutant-'+name+'.log')
  with logfile.open('w') as log:r=subprocess.run(['go','test','./internal/oracle','-run',pattern,'-v','-count=1'],stdout=log,stderr=subprocess.STDOUT)
  out=logfile.read_text();assert r.returncode!=0 and '--- FAIL: Test' in out and '[build failed]' not in out and 'clang failed' not in out,out
  if name=='record-discard-release':assert 'LeakSanitizer' in out
  print(name+': compiled execution caught',flush=True)
 finally:p.write_text(original)
