import pathlib,json,subprocess,shutil,re,time
root=pathlib.Path('review/test-audit/internal-native-split_units');menu=json.loads((root/'menu.json').read_text());probes=json.loads((root/'probes.json').read_text());originals=json.load(open('/tmp/u053/originals.json')); target=pathlib.Path('/tmp/u053/standalone-runtime');shutil.copytree('internal/native/runtime',target,dirs_exist_ok=True)
for f,s in originals.items():
 if f.startswith('internal/native/runtime/'):(target/pathlib.Path(f).name).write_text(s)
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all']
results=[]
for m in menu['mutations']+probes:
 id=m['id'];f=m['file'];s=originals[f]; start=time.monotonic()
 if id.startswith('P'):
  sig=m['signature'];mut=s.replace(sig,sig+'\n\t'+m['statement'],1)
 else:mut=s.replace(m['before'],m['after'])
 log=open('/tmp/u053/validate-'+id+'.log','w')
 if '/runtime/' in f:
  p=target/pathlib.Path(f).name;p.write_text(mut)
  source=p if p.suffix=='.c' else target/'string.c'
  cmd=['clang']+flags+['-I',str(target),'-c',str(source),'-o','/tmp/u053/validate.o']
  run=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT);p.write_text(s)
 else:
  p=pathlib.Path('/tmp/u053/overlay-'+id+'.go');p.write_text(mut);ov=pathlib.Path('/tmp/u053/overlay-'+id+'.json');ov.write_text(json.dumps({'Replace':{str(pathlib.Path(f).resolve()):str(p)}}));cmd=['go','vet','-overlay',str(ov),'./internal/native/'];run=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
 log.close();results.append(dict(id=id,returncode=run.returncode,wall=time.monotonic()-start,command=' '.join(cmd)));print(id,run.returncode,flush=True)
pathlib.Path('/tmp/u053/validation.json').write_text(json.dumps(results,indent=2))
