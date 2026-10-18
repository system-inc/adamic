from pathlib import Path
import subprocess,json,time,difflib
root=Path.cwd();out=root/'review/compiler/loom-r-fx3-refusals'; results=[]
for name,relative,test,expected in [
 ('every','internal/lower/object.go','TestArrayNarrowingEveryRefused','got <nil>'),
 ('findLast','internal/lower/object.go','TestArrayNarrowingFindLastRefused','got <nil>'),
 ('literal','internal/lower/object.go','TestFX5NumberLiteralRefused','got <nil>'),
 ('registration','internal/lower/agree_test.go','TestNativeAgreementRowsNameRegisteredPrograms','must name a program in nativeAgreementPrograms by a literal'),
]:
    original=(root/relative).read_text(); patch=(root/'review/compiler/loom-fx3-refusals-3'/(name+'.patch')).read_text()
    old=''.join(line[1:] for line in patch.splitlines(True) if line.startswith('-') and not line.startswith('---'))
    new=''.join(line[1:] for line in patch.splitlines(True) if line.startswith('+') and not line.startswith('+++'))
    assert original.count(old)==1,(name,original.count(old))
    mutated=original.replace(old,new,1)
    source=out/(name+'.go.txt');source.write_text(mutated)
    (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=relative,tofile=relative)))
    overlay=out/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/relative):str(source)}})+'\n')
    args=['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-v','-timeout','90s','-parallel','4']
    began=time.monotonic()
    with (out/('mutant-'+name+'.log')).open('w') as log:
        result=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=150)
    text=(out/('mutant-'+name+'.log')).read_text()
    assert result.returncode!=0 and expected in text,(name,text)
    results.append(dict(name=name,command=args,exit=result.returncode,catcher=test,expected_assertion=expected,seconds=round(time.monotonic()-began,3)))
    (out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(name+': caught by '+test,flush=True)
