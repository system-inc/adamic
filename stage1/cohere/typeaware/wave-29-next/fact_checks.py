#!/usr/bin/env python3
"""Raw fact and lifetime controls independent of rule mutation."""
import json,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4];source=Path(__file__).resolve().parent
d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True);data=Path(sys.argv[2]).resolve();compiler=Path('/workspace/wave29-regex-controls/adamic')
def run(name,args,expected=0):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=repo,stdout=out,stderr=err)
 assert p.returncode==expected,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
original=repo/'bridge/tsgo/checker/reference_symbol_origins.go';text=original.read_text();needle='symbol = c.GetShorthandAssignmentValueSymbol(parent)';assert text.count(needle)==1
(d/'fact-mutant.go').write_text(text.replace(needle,'symbol = c.GetSymbolAtLocation(node)'))
(d/'fact-overlay.json').write_text(json.dumps({'Replace':{str(original):str(d/'fact-mutant.go')}}))
run('fact-archive',['go','build','-buildmode=c-archive','-overlay',d/'fact-overlay.json','-o',d/'fact-mutant.a','./bridge/tsgo/archive'])
run('fact-native',[compiler,'build',source/'runner.a','-o',d/'fact-native','--tsgo',d/'fact-mutant.a'])
group=data/'group-00';args=[group/'config.json',group/'manifest',group/'profiles.frames'];got=run('fact-mutant',[d/'fact-native',*args])
assert got!=(data/'group-00-go.stdout').read_bytes() and not(d/'fact-mutant.stderr').read_bytes()
print('shorthand accessor mutant compiled, exit 0, empty stderr, killed only by byte comparison',flush=True)
(d/'anchor.a').write_text('foo;\n');(d/'config.json').write_text('{"compilerOptions":{"target":"ES2022"},"files":["anchor.a"]}')
(d/'released.a').write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,3,'Identifier','reference-symbol-origins'));\n")
run('released-build',[compiler,'build',d/'released.a','-o',d/'released','--tsgo','/workspace/wave29-next-checker.a'])
run('released',[d/'released',d/'config.json',d/'anchor.a'],expected=70)
assert 'invalid or released checker handle' in (d/'released.stderr').read_text();print('new question rejects released handle: exit 70',flush=True)
run('released-asan-build',[compiler,'build',d/'released.a','-o',d/'released-asan','--tsgo',data/'asan-checker.a','--sanitize'])
run('released-asan',[d/'released-asan',d/'config.json',d/'anchor.a'],expected=70)
error=(d/'released-asan.stderr').read_text()
assert 'invalid or released checker handle' in error and 'Sanitizer' not in error
print('instrumented released handle rejects with exit 70, no sanitizer diagnostic',flush=True)
original=repo/'bridge/tsgo/archive/main.go';text=original.read_text();needle='delete(programs.live, uint64(handle))';assert text.count(needle)==1
(d/'registry-mutant.go').write_text(text.replace(needle,'// Mutant: retain released program.'))
(d/'registry-overlay.json').write_text(json.dumps({'Replace':{str(original):str(d/'registry-mutant.go')}}))
run('registry-archive',['go','build','-buildmode=c-archive','-overlay',d/'registry-overlay.json','-o',d/'registry-mutant.a','./bridge/tsgo/archive'])
run('registry-native',[compiler,'build',d/'released.a','-o',d/'registry-native','--tsgo',d/'registry-mutant.a'])
run('registry-mutant',[d/'registry-native',d/'config.json',d/'anchor.a'])
assert not(d/'registry-mutant.stderr').read_bytes();print('retained registry mutant exit 0, empty stderr, killed by required release panic',flush=True)
