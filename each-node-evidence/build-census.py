from pathlib import Path
import sys,re,subprocess,time,json
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/each-node');source=repo/'stage3/census/latent/make_overlay.py';text=source.read_text();text=text.replace("scratch / name.replace('/', '_')", "scratch / (name.replace('/', '_') + '.txt')");text=re.sub(r"((?:scratch|output) / '[^']+\.go)'",r"\1.txt'",text);sys.argv=[str(source),str(repo),str(s/'overlay')];exec(compile(text,str(source),'exec'),{'__file__':str(source),'__name__':'__main__'});start=time.monotonic()
with (s/'census-build.log').open('w') as f:p=subprocess.run(['go','build','-buildvcs=false','-overlay='+str(s/'overlay/overlay.json'),'-o',str(s/'latent'),'./stage3/census/latent/tool'],cwd=repo,stdout=f,stderr=subprocess.STDOUT)
result=dict(exit=p.returncode,seconds=time.monotonic()-start);(s/'census-build-result.json').write_text(json.dumps(result)+'\n');print(result);raise SystemExit(p.returncode)
