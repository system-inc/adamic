"""Hold parked rule code against Go in a frozen harness copy using current-main compiler files."""
import json,os,shutil,subprocess,tarfile
from pathlib import Path
OWN=Path(__file__).resolve().parent
REPOSITORY=OWN.parents[4]
COMPILER=Path(os.environ.get('ADAMIC_PARK_COMPILER_ROOT','/workspace/adamic'))
ROOT=Path('/tmp/wave10-park-validation')
ROOT.mkdir(exist_ok=True)
slugs=['react-no-unsafe','react-self-closing-comp','sort-vars','typescript-ban-tslint-comment','typescript-no-invalid-this','arrow-body-style','max-lines','no-extra-bind','default-case','no-extra-label','no-fallthrough']
archive=ROOT/'foundation.tar'
with archive.open('wb') as out:subprocess.run(['git','archive','641b887ad10c109ba8a30f54966b8c6a87d7c579','stage1','cmd/lint-registry'],cwd=REPOSITORY,stdout=out,check=True)
with tarfile.open(archive) as source:source.extractall(ROOT,filter='data')
shutil.rmtree(ROOT/'stage1/cohere/lint/rules/structure-tailwind-no-physical-direction',ignore_errors=True)
for name in ['internal','bridge','oracle','cohere','go.mod','go.sum','go.work']:
 if not ((ROOT/name).exists() or (ROOT/name).is_symlink()): (ROOT/name).symlink_to(COMPILER/name,target_is_directory=(COMPILER/name).is_dir())
if not (ROOT/'cmd/adamic').exists(): (ROOT/'cmd/adamic').symlink_to(COMPILER/'cmd/adamic',target_is_directory=True)
for slug in slugs:
 shutil.copytree(REPOSITORY/'stage1/cohere/lint/rules'/slug,ROOT/'stage1/cohere/lint/rules'/slug,dirs_exist_ok=True)
env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/lint-wave1-10-next/typescript';env['ADAMIC_GATE_UNCACHED']='1'
def run(name,args,cwd=ROOT,extra=None):
 print('running',name,flush=True)
 with (ROOT/(name+'.log')).open('w') as log:subprocess.run(args,cwd=cwd,env=env|(extra or {}),stdout=log,stderr=subprocess.STDOUT,check=True)
 print('passed',name,flush=True)
run('prepare',['python3',str(ROOT/'stage1/cohere/lint/rules/default-case/validate.py')])
overlay=Path('/tmp/lint-wave1-10-next/overlay.json')
r=json.loads(overlay.read_text());harness=Path(r['Replace'][str(ROOT/'stage1/cohere/lint/lint_test.go')])
s=harness.read_text();needle='for _, descriptor := range prepareRegistry(t, ".") {\n\t\tvar change'
s=s.replace(needle,'for _, descriptor := range prepareRegistry(t, ".") {\n        if descriptor.Name != "react/no-unsafe" && descriptor.Name != "sort-vars" && descriptor.Name != "@typescript-eslint/ban-tslint-comment" && descriptor.Name != "@typescript-eslint/no-invalid-this" { continue }\n\t\tvar change')
s=s.replace('{"Node", node(t, directory, path, false)}, {"native",','{"Node", node(t, directory, path, false)}, {"emitted JS", emitted(t, directory, path)}, {"native",')
harness.write_text(s)
flags='-overlay='+str(overlay)
run('upstream',['go','test',flags,'-count=1','-v','-timeout','30m','./stage1/cohere/lint','-run','^Test(OriginalUpstream|Wave10Upstream|NextUpstream|FourthUpstream)$'])
# Refresh independent Go fixture answers after capturing the real upstream tests.
cohere=COMPILER/'cohere';virtual=str(cohere/'adamic_park_snapshot.go');side=str(ROOT/'stage1/cohere/lint/rules/react-self-closing-comp/snapshot.go.txt')
snapshot_overlay=ROOT/'snapshot-overlay.json';snapshot_overlay.write_text(json.dumps({'Replace':{virtual:side}}))
run('snapshot-go',['go','run','-overlay='+str(snapshot_overlay),virtual,'/tmp/lint-wave1-10-next/captured.json',str(ROOT/'stage1/cohere/lint/rules/react-self-closing-comp/snapshot.a'),'/tmp/wave10-selfclosing-want.txt'],cwd=cohere)
virtual=str(cohere/'adamic_park_repairs.go');side=str(ROOT/'stage1/cohere/lint/rules/arrow-body-style/repairs.go.txt')
repair_overlay=ROOT/'repairs-overlay.json';repair_overlay.write_text(json.dumps({'Replace':{virtual:side}}))
# Preserve the three documented independent-parser exclusions used by TestNextUpstream.
records=json.loads(Path('/tmp/lint-wave1-10-next/captured.json').read_text())
exclusions=json.loads((ROOT/'stage1/cohere/lint/rules/max-lines/testdata/parser-exclusions.json').read_text())
repair_records=ROOT/'repair-records.json'
repair_records.write_text(json.dumps({key:value for key,value in records.items() if key not in exclusions}))
run('repairs-go',['go','run','-overlay='+str(repair_overlay),virtual,str(repair_records),str(ROOT/'stage1/cohere/lint/rules/arrow-body-style/repairs.a'),str(ROOT/'stage1/cohere/lint/rules/arrow-body-style/testdata/repairs-answer.txt')],cwd=cohere)
pattern='^Test(Original(Witnesses|Corpus|SelfSnapshots)|Wave10(Shapes|Corpus|Unterminated)|Next(Witnesses|Corpus|Mutants|Repairs)|Fourth(Witnesses|Corpus|Mutants|Shapes|PatternRefusal)|Mutants)$'
run('oracle',['go','test',flags,'-count=1','-v','-timeout','30m','./stage1/cohere/lint','-run',pattern])
run('numeric-listeners',['python3',str(ROOT/'stage1/cohere/lint/rules/default-case/listeners_validate.py')],extra={'ADAMIC_LISTENER_COMPILER_ROOT':str(COMPILER)})
run('vet',['go','vet',flags,'./stage1/cohere/lint'])
print('PARKING ORACLE PASS. Raw logs:',ROOT,flush=True)
