#!/usr/bin/env python3
"""Mutate named contracts with overlays so concurrent validations see stable sources."""
from pathlib import Path
import os, subprocess, json
root=Path('/workspace/adamic');source=root/'internal/lower/records_named.go';logs=Path(os.environ.get('NAMED_INDEX_MUTANT_LOGS','/tmp/named-index-mutants'));logs.mkdir(parents=True,exist_ok=True);original=source.read_text()
mutants=[('index-read-type','recordReadType','return read, nil','TestNamedRecordReadTypes','named read used index type'),('unrestricted-write','recordNamedWrite','return nil','TestNamedRecordRefusals','want NotYet named property named'),('erase-named-contract','sameRecordNamedContracts','return true','TestNamedRecordRefusals','want NotYet seen as')]
for name,function,replacement,test,witness in mutants:
 at=original.index('func (l *lowering) '+function+'(');insert=original.index('\n',at)+1
 modified=logs/(name+'.go.txt');modified.write_text(original[:insert]+'\t'+replacement+' // deliberate mutant\n'+original[insert:]);overlay=logs/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(source):str(modified)}}))
 with (logs/(name+'.log')).open('w') as output:result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-timeout','85s'],cwd=root,stdout=output,stderr=subprocess.STDOUT,timeout=89)
 observed=(logs/(name+'.log')).read_text();assert result.returncode!=0 and witness in observed,(name,result.returncode,observed);print(name+': caught by '+test+' (exit '+str(result.returncode)+')',flush=True)
assert source.read_text()==original
