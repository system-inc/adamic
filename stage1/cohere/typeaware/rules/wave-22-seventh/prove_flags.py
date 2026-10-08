#!/usr/bin/env python3
"""Prove the numeric per-file guards can fail the external byte oracle."""
import pathlib,subprocess,time,json
ROOT=pathlib.Path(__file__).resolve().parents[5]
WORK=pathlib.Path('/workspace/wave-22-seventh-work')
ENTRY=WORK/'source'/pathlib.Path(__file__).resolve().parent.relative_to(ROOT)/'driver.a'
COMPILER='/workspace/wave-22-seventh-adamic'
def run(label,args):
    before=time.perf_counter()
    with (WORK/(label+'.stdout')).open('wb') as out,(WORK/(label+'.stderr')).open('wb') as err:
        result=subprocess.run([str(a) for a in args],stdout=out,stderr=err,cwd=ROOT)
    (WORK/(label+'.json')).write_text(json.dumps({'command':[str(a) for a in args],'exit':result.returncode,'seconds':time.perf_counter()-before})+'\n')
    if result.returncode:raise RuntimeError(label+': '+(WORK/(label+'.stderr')).read_text())
truth=(WORK/'controls-go.stdout').read_bytes()
for name,relative,before,after in [
    ('computed-trivia','no-object-type-as-default-prop/rule.a','this.trim(c.fullRaw(inner))','c.raw(inner).trim()'),
    ('jsx-attribute-flag','common/context.a','this.hasJsxAttributes=true;','this.hasJsxAttributes=false;'),
    ('class-fallback-flag','no-unstable-nested-components/rule.a','if(!this.components.hasReactClass){return false;}','if(!this.components.hasReactClass){return true;}')]:
    target=ENTRY.parent/relative;original=target.read_text();assert original.count(before)==1;target.write_text(original.replace(before,after))
    try:
        run(name+'-build',[COMPILER,'build',ENTRY,'-o',WORK/name,'--tsgo','/workspace/wave-22-seventh-checker.a'])
        run(name+'-run',[WORK/name,WORK/'tsconfig.json',WORK/'controls.manifest'])
        result=(WORK/(name+'-run.stdout')).read_bytes()
        if result==truth or (WORK/(name+'-run.stderr')).read_bytes():raise RuntimeError(name+' survived or crashed')
        position=next((i for i,(a,b) in enumerate(zip(result,truth)) if a!=b),min(len(result),len(truth)))
        print(name,'byte-only mutant caught at',position,flush=True)
    finally:target.write_text(original)
