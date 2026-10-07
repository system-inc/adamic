"""Raw numeric-question and stale-handle mutation checks, separate from rule policy."""
import pathlib,subprocess,json,time,argparse
OWN=pathlib.Path(__file__).resolve().parent;ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);args=p.parse_args();S=pathlib.Path(args.scratch);records=[]
def run(name,command,expected=0):
 start=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,command)),cwd=ROOT,stdout=out,stderr=err)
 records.append(dict(name=name,command=list(map(str,command)),exit=r.returncode,seconds=time.monotonic()-start));(S/'bridge-runs.json').write_text(json.dumps(records,indent=2)+'\n');assert r.returncode==expected,(name,r.returncode,(S/(name+'.stderr')).read_text());return (S/(name+'.stdout')).read_bytes()
probe=S/'released.a';probe.write_text("import {panic,programArguments,readTextFile,tsgoProgram,tsgoRelease} from 'adamic';import {byteOffsets} from '"+str(ROOT/'stage1/cohere/typeaware/unary_minus.ts')+"';import {SyntaxFacts} from '"+str(OWN/'jsx_syntax_facts.a')+"';const args=programArguments();const path=args[1]??panic('missing source');const source=readTextFile(path);if(source.kind==='Error'){panic(source.message);}const offsets=byteOffsets(source.text);const p=tsgoProgram(args[0]??panic('missing config'),[path]);tsgoRelease(p);const facts=new SyntaxFacts(p,path,offsets[source.text.length]??panic('missing end'));console.log(facts.sourceCount.toString());\n")
exe=S/'released';run('released-build',[S/'adamic','build',probe,'-o',exe,'--tsgo',S/'checker.a']);file=(S/'controls.manifest').read_text().splitlines()[0];config=S/'tsconfig.json';run('released-run',[exe,config,file],70);assert (S/'released-run.stderr').read_bytes()==b'adamic: panic: invalid or released checker handle\n'
for name,path,old,new in [('numeric-question','bridge/tsgo/checker/jsx_syntax_facts.go','records.number(uint64(n.Kind))','records.number(uint64(n.Kind)+1)'),('stale-registry','bridge/tsgo/archive/main.go','delete(programs.live, uint64(handle))','// Mutant keeps the stale handle live.')]:
 original=ROOT/path;text=original.read_text();assert text.count(old)==1;mutant=S/(name+'.go');mutant.write_text(text.replace(old,new));overlay=S/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));archive=S/(name+'.a');run(name+'-archive',['go','build','-buildmode=c-archive','-overlay',overlay,'-o',archive,'./bridge/tsgo/archive']);binary=S/(name+'-mutant');entry=OWN/'main.a' if name=='numeric-question' else probe;run(name+'-build',[S/'adamic','build',entry,'-o',binary,'--tsgo',archive])
 output=run(name+'-run',[binary,config,S/'controls.manifest'] if name=='numeric-question' else [binary,config,file]);assert not (S/(name+'-run.stderr')).read_bytes()
 if name=='numeric-question':
  truth=(S/'controls-go.stdout').read_bytes();assert output!=truth;at=next((i for i,(a,b) in enumerate(zip(output,truth)) if a!=b),min(len(output),len(truth)));print('numeric question: compiled, exit 0, empty stderr; Go byte comparison catches byte '+str(at),flush=True)
 else:print('stale registry: compiled, exit 0, empty stderr; required exit-70 contract catches mutant',flush=True)
 binary.unlink();archive.unlink()
print('PASS: numeric query mutation and released-handle contract',flush=True)
