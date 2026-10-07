from pathlib import Path
import subprocess,json,re
root=Path('/workspace/adamic');out=Path('/workspace/wave-23/landing-b8fb/jsx-next');base=root/'stage1/cohere/typeaware';stage=out.parent/'behavior/adamic';obs={}
def run(label,args):
 p=subprocess.run([str(x) for x in args],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 (out/(label+'.stdout')).write_bytes(p.stdout);(out/(label+'.stderr')).write_bytes(p.stderr);return p
args=['/workspace/wave-23/react-blocker/tsconfig.json','/workspace/wave-23/jsx-next/manifest']
expected=run('go-listeners',[out/'go-oracle',*args,'--listener-kinds']);assert expected.returncode==0,expected.stderr
p=run('go-findings',[out/'go-oracle',*args]);assert p.returncode==0 and p.stdout==Path('/workspace/wave-23/jsx-next/findings.stdout').read_bytes()
for label,flags in [('native',[]),('native-asan',['--sanitize'])]:
 p=run(label+'-build',[stage,'build',base/'wave_23_jsx_listener_probe.a','-o',out/label,*flags]);assert p.returncode==0,p.stderr
 p=run(label,[out/label]);assert p.returncode==0 and not p.stderr and p.stdout==expected.stdout
 obs[label]={'exit':p.returncode,'bytes':len(p.stdout)}
files=['jsx_fragments_listener.a','jsx_no_constructed_context_values_listener.a','jsx_no_undef_listener.a']
for i,f in enumerate(files):
 d=out/('mutant-'+str(i));d.mkdir(exist_ok=True);s=(base/f).read_text().replace("'./wave_23_listener_kinds.a'",repr(str(base/'wave_23_listener_kinds.a')));s=re.sub(r'kinds: \[\d+', 'kinds: [0',s);(d/f).write_text(s)
 probe=(base/'wave_23_jsx_listener_probe.a').read_text()
 for j,g in enumerate(files):probe=probe.replace("'./"+g+"'",repr(str((d if i==j else base)/g)))
 (d/'probe.a').write_text(probe)
 p=run('mutant-'+str(i)+'-build',[stage,'build',d/'probe.a','-o',d/'native']);assert p.returncode==0,p.stderr
 p=run('mutant-'+str(i),[d/'native']);assert p.returncode==0 and not p.stderr and p.stdout!=expected.stdout
 obs[f+' metadata mutant']={'exit':p.returncode,'first_difference':next(k for k,(a,b) in enumerate(zip(p.stdout,expected.stdout)) if a!=b)}
p=run('parser-build',[stage,'build',base/'wave_23_react_parser_probe.a','-o',out/'parser','--tsgo',out.parent/'behavior/checker.a']);assert p.returncode==0,p.stderr
for f in ['fragment','context','undef']:
 p=run(f+'-parser',[out/'parser',args[0],'/workspace/wave-23/jsx-next/'+f+'.tsx']);assert p.returncode==70
 obs[f+' JSX probe']={'exit':p.returncode,'stderr':p.stderr.decode()}
(out/'observations.json').write_text(json.dumps(obs,indent=2)+'\n');print(json.dumps(obs,indent=2))
