#!/usr/bin/env python3
"""Released-question probes and the unmodified shared-parser dependency."""
import pathlib,subprocess,time,json,statistics
r=pathlib.Path(__file__).resolve().parents[5];w=pathlib.Path('/workspace/wave-22-sixth-work')
def run(label,args):
 start=time.perf_counter()
 with (w/(label+'.stdout')).open('wb') as out,(w/(label+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=r,stdout=out,stderr=err)
 result={'command':[str(x) for x in args],'exit':p.returncode,'seconds':time.perf_counter()-start};(w/(label+'.json')).write_text(json.dumps(result)+'\n');return result
probe=w/'released-input.ts';probe.write_text('x;\n')
for q in ['symbol-declaration-syntax','node-symbol-details']:
 result=run('released-'+q,[w/'released',w/'tsconfig.json',probe,q]);assert result['exit']==70,result
 err=(w/('released-'+q+'.stderr')).read_text();assert 'released' in err or 'invalid program' in err,err
print('both released questions refused with exit 70',flush=True)
result=run('main-parser-build',['/workspace/wave-22-sixth-adamic','build',r/'stage1/cohere/typeaware/rules/wave-22-sixth/driver.a','-o',w/'native-main','--tsgo','/workspace/wave-22-sixth-checker.a'])
assert result['exit']==1 and 'escaping a constructor' in (w/'main-parser-build.stderr').read_text(),result
print('unmodified main parser constructor dependency reproduced; no shared edits',flush=True)
