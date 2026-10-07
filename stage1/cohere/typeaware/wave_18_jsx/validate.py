"""Independent production Go comparison for the three numeric-node JSX rules."""
import argparse, pathlib, subprocess, os, json, shutil, time, statistics, re
OWN=pathlib.Path(__file__).resolve().parent
ROOT=OWN.parents[3]
p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);args=p.parse_args();S=pathlib.Path(args.scratch);S.mkdir(parents=True,exist_ok=True);runs=[]
def run(name,command,exit=0,env=None):
 start=time.monotonic()
 with (S/(name+'.stdout')).open('wb') as out,(S/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,command)),cwd=ROOT,stdout=out,stderr=err,env=env)
 runs.append(dict(name=name,command=list(map(str,command)),exit=r.returncode,seconds=time.monotonic()-start));(S/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
 assert r.returncode==exit,(name,r.returncode,(S/(name+'.stderr')).read_text())
 return (S/(name+'.stdout')).read_bytes()
def compare(name,config,manifest,flags=()):
 truth=run(name+'-go',[S/'oracle',config,manifest,*flags]);actual=run(name+'-native',[S/'native',config,manifest,*flags]);assert actual==truth,(name,'native byte mismatch');assert not (S/(name+'-native.stderr')).read_bytes()
 sanitized=run(name+'-asan',[S/'native-asan',config,manifest,*flags]);assert sanitized==truth,(name,'sanitizer byte mismatch');assert not (S/(name+'-asan.stderr')).read_bytes();print(name+': identical '+str(len(truth))+' bytes; '+truth.decode().splitlines()[-1],flush=True);return truth
if not (S/'native-asan').exists():
 config=S/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','module':'NodeNext','jsx':'preserve','types':[],'moduleDetection':'force'},'files':['anchor.d.ts']}));(S/'anchor.d.ts').write_text('declare module "react" {export const Fragment:any;export function createElement(...args:any[]):any;}\n')
 paths=[]
 for i,source in enumerate(json.loads((OWN/'controls.json').read_text())):
  path=S/('control-'+str(i).zfill(3)+'.tsx');path.write_text(source);paths.append(str(path))
 all_manifest=S/'all-controls.manifest';all_manifest.write_text('\n'.join(paths)+'\n')
 overlay=S/'overlay.json';virtual=ROOT/'cohere/adamic_wave18jsx_oracle.go';overlay.write_text(json.dumps({'Replace':{str(virtual):str(OWN/'testdata/oracle.go')}}))
 run('stage0-build',['go','build','-o',S/'adamic','./cmd/adamic'])
 run('oracle-build',['go','-C',ROOT/'cohere','build','-overlay',overlay,'-o',S/'oracle',virtual])
 (S/'controls.manifest').write_bytes(run('filter',[S/'oracle',config,all_manifest,'--valid-sources']))
 run('checker-build',['go','build','-buildmode=c-archive','-o',S/'checker.a','./bridge/tsgo/archive'])
 run('native-build',[S/'adamic','build',OWN/'main.a','-o',S/'native','--tsgo',S/'checker.a'])
 env=dict(os.environ,CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
 run('checker-asan-build',['go','build','-buildmode=c-archive','-o',S/'checker-asan.a','./bridge/tsgo/archive'],env=env)
 run('native-asan-build',[S/'adamic','build',OWN/'main.a','-o',S/'native-asan','--tsgo',S/'checker-asan.a','--sanitize'])
config=S/'tsconfig.json';manifest=S/'controls.manifest'
truth=compare('controls',config,manifest)
for name,flags in [('element',['--element']),('allow-globals',['--allow-globals']),('both-options',['--element','--allow-globals'])]:compare(name,config,manifest,flags)
for name,folder,old,new in [('fragments','jsx_fragments',"name.text === 'Fragment'","name.text !== 'Fragment'"),('undef','jsx_no_undef','if(reference.resolved)','if(!reference.resolved)'),('adjacent','no_adjacent_inline_elements','if(previous && current)','if(previous && !current)')]:
 directory=S/(name+'-source');shutil.copytree(OWN,directory,dirs_exist_ok=True)
 for copied in directory.rglob('*.a'):
  original=OWN/copied.relative_to(directory)
  copied.write_text(re.sub(r"(['\"])([^'\"]+\.ts)\1",lambda m:repr(str((original.parent/m.group(2)).resolve())),copied.read_text()))
 path=directory/folder/'index.a';text=path.read_text();assert text.count(old)==1,(name,old);path.write_text(text.replace(old,new));exe=S/(name+'-mutant');run(name+'-mutant-build',[S/'adamic','build',directory/'main.a','-o',exe,'--tsgo',S/'checker.a']);output=run(name+'-mutant-run',[exe,config,manifest]);assert not (S/(name+'-mutant-run.stderr')).read_bytes();assert output!=truth,(name,'survived');at=next((i for i,(a,b) in enumerate(zip(output,truth)) if a!=b),min(len(output),len(truth)));print(name+': compiled, exit 0, empty stderr; Go byte comparison catches byte '+str(at),flush=True);exe.unlink()
for name,env_config,env_manifest in [('compiler','ADAMIC_WAVE18_COMPILER_CONFIG','ADAMIC_WAVE18_COMPILER_MANIFEST'),('repository','ADAMIC_WAVE18_REPOSITORY_CONFIG','ADAMIC_WAVE18_REPOSITORY_MANIFEST')]:
 corpus_config=os.environ.get(env_config);corpus_manifest=os.environ.get(env_manifest)
 if not corpus_config or not corpus_manifest:continue
 compare(name,corpus_config,corpus_manifest)
 timed={'native':[],'go':[]}
 for round_ in range(3):
  for backend in (['native','go'] if round_%2==0 else ['go','native']):
   run(name+'-'+str(round_)+'-'+backend+'-time',[S/('oracle' if backend=='go' else 'native'),corpus_config,corpus_manifest,'--count']);timed[backend].append(runs[-1]['seconds'])
 print(name+' median process seconds native='+str(statistics.median(timed['native']))+' Go='+str(statistics.median(timed['go'])),flush=True)
(S/'native-mutants.json').write_text(json.dumps([r for r in runs if 'mutant' in r['name']],indent=2)+'\n')
print('PASS: complete records, options, byte-only rule mutants, corpora and sanitizers',flush=True)
