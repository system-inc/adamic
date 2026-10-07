#!/usr/bin/env python3
"""Positive production Go findings, working JSX parsing and missing checker integration."""
import json,os,subprocess,sys
from pathlib import Path
s=Path(__file__).resolve().parent;r=s.parents[3];d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
c=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-latest-kernel-adamic'));commands=[]
def run(name,args,cwd=r,expected=0):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode));(d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 if expected is not None:assert p.returncode==expected,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
v=r/'cohere/adamic_wave29_fourth_source_oracle.go';(d/'overlay.json').write_text(json.dumps({'Replace':{str(v):str(s/'testdata/source_oracle.go')}}))
run('go-build',['go','build','-overlay',d/'overlay.json','-o',d/'oracle',v],cwd=r/'cohere')
paths=[];profile=[]
for name in ['react-jsx-fragments','react-jsx-no-constructed-context-values','react-jsx-no-undef']:
 probe=s/'rules'/name/'gaps/source.a';code=probe.read_text().split("new Parser('")[1].split("', 'input.tsx')")[0]
 path=d/(name+'.tsx');path.write_text(code+'\n');paths.append(str(path));profile.append(dict(rule=name.replace('react-','react/',1)))
(d/'manifest').write_text('\n'.join(paths)+'\n');(d/'profile.json').write_text(json.dumps(profile));(d/'config.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'jsx':'preserve','module':'ESNext'},'files':paths}))
truth=run('go',[d/'oracle',d/'config.json',d/'manifest',d/'profile.json']);assert truth.splitlines()[-1]==b'findings 3'
for row in profile:assert ('\t'+row['rule']+'\t').encode() in truth
print('Go production callbacks: three real JSX sources, three findings, zero fixes/suggestions',flush=True)
for row in profile:
 name=row['rule'].replace('/','-');probe=s/'rules'/name/'gaps/source.a'
 run(name+'-build',[c,'build',probe,'-o',d/name]);got=run(name,[d/name],expected=None);status=commands[-1]['exit'];err=(d/(name+'.stderr')).read_text()
 assert status==0 and not err and b'Jsx' in got and b'TypeAssertionExpression' not in got
 print(name+': compiled, exit 0, empty stderr, real JSX nodes; parser blocker closed, checker source adapter still required',flush=True)
