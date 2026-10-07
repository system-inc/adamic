"""Native numeric listener subscriptions versus the pinned Go AST enum."""
import argparse,json,pathlib,re,subprocess
parser=argparse.ArgumentParser()
for name in ['artifacts','stage0','archive','asan-archive']:parser.add_argument('--'+name,required=True)
args=parser.parse_args();source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];out=pathlib.Path(args.artifacts).resolve();out.mkdir(parents=True,exist_ok=True)
def run(name,command,cwd=repo):
 with open(out/(name+'.stdout'),'wb') as stdout,open(out/(name+'.stderr'),'wb') as stderr:p=subprocess.run([str(x) for x in command],cwd=cwd,stdout=stdout,stderr=stderr)
 result=(out/(name+'.stdout')).read_bytes();errors=(out/(name+'.stderr')).read_bytes();assert p.returncode==0,(name,p.returncode,errors);return result,errors
virtual=repo/'cohere/adamic_wave08_listener_oracle.go';(out/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/listener_oracle.go')}}))
truth,errors=run('go',['go','run','-overlay',out/'overlay.json',virtual],repo/'cohere');assert errors==b''
run('build',[args.stage0,'build',source/'testdata/listener_probe.a','-o',out/'native','--tsgo',args.asan_archive,'--sanitize']);actual,errors=run('native',[out/'native']);assert actual==truth and errors==b''
p=source.parent/'no_require_imports.a';text=p.read_text();assert text.count('[214, 284]')==1;(out/'mutant.a').write_text(text.replace('[214, 284]','[215, 284]').replace("'./","'"+str(p.parent)+"/"))
probe=source/'testdata/listener_probe.a';text=re.sub(r"from '([^']+)'",lambda m:"from '"+str((probe.parent/m.group(1)).resolve())+"'",probe.read_text()).replace(str(p),str(out/'mutant.a'));(out/'mutant-probe.a').write_text(text)
run('mutant-build',[args.stage0,'build',out/'mutant-probe.a','-o',out/'mutant','--tsgo',args.archive]);actual,errors=run('mutant',[out/'mutant']);assert actual!=truth and errors==b''
print('PASS nine numeric listener declarations match Go under sanitizers; compiling exit-zero CallExpression 214 -> 215 mutant caught only by Go kind bytes',flush=True)
