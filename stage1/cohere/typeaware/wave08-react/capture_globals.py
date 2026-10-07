"""Capture all upstream Globals calls while retaining their assertions."""
import argparse,json,os,pathlib,subprocess
p=argparse.ArgumentParser();p.add_argument('directory');a=p.parse_args()
s=pathlib.Path(__file__).resolve().parent;r=s.parents[3];out=pathlib.Path(a.directory).resolve();out.mkdir(parents=True,exist_ok=True)
original=r/'cohere/internal/lint/rules/react/globals_test.go'
replacement=out/'globals_test.go'
replacement.write_text(original.read_text().replace('rule_testing.RunTyped(', 'wave08Capture(').replace('rule_testing.Run(', 'wave08CapturePlain('))
virtual=r/'cohere/internal/lint/rules/react/adamic_wave08_capture_test.go'
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(replacement),str(virtual):str(s/'testdata/capture_globals_test.go')}}))
with (out/'capture.stdout').open('wb') as stdout,(out/'capture.stderr').open('wb') as stderr:
 result=subprocess.run(['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestGlobals','-count=1','-v'],cwd=r/'cohere',stdout=stdout,stderr=stderr,env=dict(os.environ,ADAMIC_WAVE08_CAPTURE=str(out/'inputs')))
assert result.returncode==0,'upstream assertions failed; see capture logs'
for path in sorted((out/'inputs').glob('*.json')):
 data=json.loads(path.read_text());folder=out/'cases'/path.stem;folder.mkdir(parents=True,exist_ok=True)
 for name,text in data['Files'].items():
  target=folder/name.lstrip('/');target.parent.mkdir(parents=True,exist_ok=True);target.write_text(text)
 (folder/'roots.manifest').write_text(''.join(str(folder/name.lstrip('/'))+'\n' for name in sorted(data['Files']) if name.endswith(('.ts','.tsx')) and not name.endswith('.d.ts')))
 (folder/'tsconfig.json').write_text(data['Config'])
 (folder/'metadata.json').write_text(json.dumps({k:v for k,v in data.items() if k!='Files'},indent=2))
print('captured',len(list((out/'inputs').glob('*.json'))),'upstream programs',flush=True)
