#!/usr/bin/env python3
"""Raw regex-program fact mutation and released-handle check."""
import json,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4];source=Path(__file__).resolve().parent;d=Path(sys.argv[1]);d.mkdir(exist_ok=True)
def run(name,args,expected=0):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=repo,stdout=out,stderr=err)
 assert p.returncode==expected,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
vm=Path('/workspace/wave29-regex-vm');old=Path('/workspace/wave29-regex-controls');compiler=old/'adamic'
original=repo/'bridge/tsgo/checker/regexp_program.go';text=original.read_text();needle='out.number(uint64(value))';assert text.count(needle)==1
(d/'regexp_program_mutant.go').write_text(text.replace(needle,"if value == 95 { value++ }; "+needle))
(d/'overlay.json').write_text(json.dumps({'Replace':{str(original):str(d/'regexp_program_mutant.go')}}))
run('fact-build',['go','build','-buildmode=c-archive','-overlay',d/'overlay.json','-o',d/'mutant.a','./bridge/tsgo/archive'])
run('native-build',[compiler,'build',source/'regex_runner.a','-o',d/'native','--tsgo',d/'mutant.a'])
got=run('fact-mutant',[d/'native',vm/'config.json',vm/'anchor.a',vm/'inputs.frames'])
assert got!=(vm/'go.stdout').read_bytes() and not(d/'fact-mutant.stderr').read_bytes();print('regex rune fact mutant compiled, exit 0, empty stderr, caught only by bytes')
(d/'released.a').write_text("import {programArguments,readTextFile,panic,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??'';const source=readTextFile(path);if(source.kind==='Error'){panic(source.message);}const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,source.text.length,'SourceFile','regexp-program\\n^x$'));")
run('released-build',[compiler,'build',d/'released.a','-o',d/'released','--tsgo',old/'checker.a'])
run('released',[d/'released',vm/'config.json',vm/'anchor.a'],expected=70)
assert 'invalid or released checker handle' in (d/'released.stderr').read_text();print('new regex question rejects released handle with exit 70')
run('released-mutant-build',[compiler,'build',d/'released.a','-o',d/'released-mutant','--tsgo','/workspace/wave29-regex-default/released-registry.a'])
run('released-mutant',[d/'released-mutant',vm/'config.json',vm/'anchor.a'])
assert not(d/'released-mutant.stderr').read_bytes();print('retained registry mutant exits 0, caught by required release panic')
