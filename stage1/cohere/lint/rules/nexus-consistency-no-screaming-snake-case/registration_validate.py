import argparse,json,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];LINT=ROOT/'stage1/cohere/lint'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
run('generate',['go','run','./cmd/lint-registry'])
replace={};files=[]
for slug,source in [('oracle',LINT/'testdata/oracle.go'),('registry',LINT/'.generated/registry.go')]+[(directory.name.replace('-','_'),directory/'oracle.go') for directory in sorted((LINT/'rules').iterdir()) if (directory/'rule.json').exists()]:
 virtual=ROOT/('cohere/adamic_nexus_'+slug+'.go');replace[str(virtual)]=str(source);files.append(virtual)
overlay=S/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}));run('Go-build',['go','build','-overlay='+str(overlay),'-o',S/'oracle',*files],ROOT/'cohere')
rows=[];number=0
for scratch in ['/tmp/wave11-screaming','/tmp/wave11-forbidden','/tmp/wave11-alias']:
 for row in json.loads((Path(scratch)/'upstream.json').read_text()):
  directory=S/('case-'+str(number));number+=1;path=directory/row['name'].lstrip('/');path.parent.mkdir(parents=True,exist_ok=True);path.write_text(row['source']);options=row['options']
  if options:
   value=json.loads(options)
   if value.get('repositoryRoot'):value['repositoryRoot']=str(directory/value['repositoryRoot'].lstrip('/'))
   options=json.dumps(value,separators=(',',':'))
  rows.append(str(path)+'\t'+row['rule']+'\t\t\tfalse\t'+options)
manifest=S/'manifest.txt';manifest.write_text('\n'.join(rows)+'\n');want=run('Go',[S/'oracle','--manifest',manifest])
run('build',['go','run',HERE/'validation_build.go',LINT/'main.ts',S/'native',S/'emitted.mjs'])
for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',LINT/'main.ts','--manifest',manifest]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',S/'emitted.mjs','--manifest',manifest]),('native',[S/'native','--manifest',manifest])]:
 actual=run(backend,cmd);assert actual==want,backend;print(backend,'equal registered cases',len(rows),'bytes',len(want),flush=True)
