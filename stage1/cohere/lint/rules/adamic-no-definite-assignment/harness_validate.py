import argparse,json,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4];LINT=ROOT/'stage1/cohere/lint'
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--allow-parser-refusals',action='store_true');a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
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
for slug in ['adamic-no-definite-assignment','base-security-require-context-access','nexus-consistency-no-screaming-snake-case','nexus-import-no-forbidden-source','nexus-import-require-path-alias']:
 scratch=Path('/tmp/wave11-parking-oracles')/slug
 for row in json.loads((Path(scratch)/'upstream.json').read_text()):
  directory=S/('case-'+str(number));number+=1;path=directory/row['name'].lstrip('/');path.parent.mkdir(parents=True,exist_ok=True);path.write_text(row['source']);options=row['options']
  if options:
   value=json.loads(options)
   if value.get('repositoryRoot'):value['repositoryRoot']=str(directory/value['repositoryRoot'].lstrip('/'))
   options=json.dumps(value,separators=(',',':'))
  rows.append(str(path)+'\t'+row['rule']+'\t\t\tfalse\t'+options)
manifest=S/'manifest.txt';manifest.write_text('\n'.join(rows)+'\n')
flags=run('Go-diagnostics',[S/'oracle','--manifest',manifest,'--diagnostics']).decode().split();assert len(flags)==len(rows)
rows=[row+'\trecovery' if flag=='1' else row+'\t' for row,flag in zip(rows,flags)]
manifest.write_text('\n'.join(rows)+'\n');want=run('Go',[S/'oracle','--manifest',manifest])
run('build',['go','run',HERE.parent/'nexus-consistency-no-screaming-snake-case/validation_build.go',LINT/'main.ts',S/'native',S/'emitted.mjs'])
if a.allow_parser_refusals:
 refused=[]
 while True:
  with (S/'parser-probe.log').open('wb') as out,(S/'parser-probe.stderr').open('wb') as err:
   probe=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(ROOT/'oracle/node.mjs'),str(LINT/'main.ts'),'--manifest',str(manifest)],cwd=ROOT,stdout=out,stderr=err)
  if probe.returncode==0:
   assert not (S/'parser-probe.stderr').read_bytes();break
  stderr=(S/'parser-probe.stderr').read_text();assert probe.returncode==70 and stderr.startswith('adamic: panic: parser slice unsupported '),stderr
  number=int([line for line in (S/'parser-probe.log').read_text().splitlines() if line.startswith('case ')][-1].split()[1]);row=rows.pop(number)
  refused.append({'manifest':row,'source':Path(row.split('\t')[0]).read_text(),'stderr':stderr});print('EXPLICIT parser refusal',len(refused),stderr.strip(),flush=True)
  manifest.write_text('\n'.join(rows)+'\n')
 (S/'parser-refusals.json').write_text(json.dumps(refused,indent=2));want=run('Go-covered',[S/'oracle','--manifest',manifest])
for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',LINT/'main.ts','--manifest',manifest]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',S/'emitted.mjs','--manifest',manifest]),('native',[S/'native','--manifest',manifest])]:
 actual=run(backend,cmd);assert actual==want,backend+' differs; retained outputs are '+str(S);print(backend,'equal registered cases',len(rows),'bytes',len(want),flush=True)
print("PASS registered raw-source comparison",flush=True)
