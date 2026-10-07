"""Each owned raw question is proven by an independently compiled Go test mutant."""
import json,pathlib,subprocess,time
ROOT=pathlib.Path(__file__).resolve().parents[4];S=pathlib.Path('/workspace/wave-26-attributes');records=[]
for name,test,needle,replacement,expected in [('jsx_structure','TestJSXStructureRoles','{name, value, tag, attrs, opening, body}','{name, value, attrs, tag, opening, body}','tag and attributes role mismatch'),('symbol_locations','TestSymbolLocationsMergedAndDestructured','declarations = symbol.Declarations','declarations = symbol.Declarations[:1]','missing declarations')]:
    original=ROOT/('bridge/tsgo/checker/'+name+'.go');text=original.read_text();assert text.count(needle)==1
    mutant=S/(name+'-mutant.go');mutant.write_text(text.replace(needle,replacement));overlay=S/(name+'-mutant.json');overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}))
    command=['go','test','-overlay',str(overlay),'./bridge/tsgo/checker','-run','^'+test+'$','-count=1','-v']
    started=time.monotonic()
    with (S/(name+'-mutant.log')).open('wb') as out:result=subprocess.run(command,cwd=ROOT,stdout=out,stderr=subprocess.STDOUT)
    records.append(dict(name=name,command=command,exit=result.returncode,seconds=time.monotonic()-started));(S/'question-mutant-runs.json').write_text(json.dumps(records,indent=2)+'\n')
    assert result.returncode==1 and expected in (S/(name+'-mutant.log')).read_text()
    print(name+': compiled, intended contract assertion caught mutant',flush=True)
