#!/usr/bin/env python3
"""Preserve outputs, hashes and command statuses; omit executables and archives."""
import gzip, hashlib, json, pathlib, shutil
ROOT=pathlib.Path(__file__).resolve().parents[5]
WORK=pathlib.Path('/workspace/wave-22-sixth-work')
DEST=ROOT/'stage1/cohere/typeaware/validation-wave-22-sixth'
DEST.mkdir(exist_ok=True)
records=[]
def preserve(path,relative):
    output=DEST/relative;output.parent.mkdir(parents=True,exist_ok=True);data=path.read_bytes()
    if path.suffix=='.stdout':output=pathlib.Path(str(output)+'.gz');output.write_bytes(gzip.compress(data,mtime=0))
    else:output.write_bytes(data)
    records.append({'path':str(output.relative_to(DEST)),'raw_bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()})
for p in WORK.iterdir():
    if p.is_file() and (p.suffix in ('.stdout','.stderr','.json','.manifest') or p.name in ('before-rebase-tip','linear-rebase-todo')):
        preserve(p,'sixth/'+p.name)
controls={p.name:p.read_text() for p in WORK.glob('control-*.tsx')}
(DEST/'controls.json').write_text(json.dumps(controls,ensure_ascii=False,indent=2)+'\n')
for batch in ['original','next','third','fourth']:
    folder=pathlib.Path('/workspace/wave-22-sixth-landing-'+batch)
    if folder.exists():
        for p in folder.iterdir():
            if p.is_file() and p.suffix in ('.stdout','.stderr','.json','.manifest'):preserve(p,'landing-'+batch+'/'+p.name)
for label in ['complete','bridge','question','question-mutant','final-checks','benchmark','upstream','cohere','vet','fetch','rebase','ancestry','landing-oracle','stage0','archive','released-build','node','node-mutant']:
    p=pathlib.Path('/workspace/wave-22-sixth-'+label+'.log')
    if p.exists():preserve(p,'logs/'+label+'.log')
truth=(WORK/'controls-go.stdout').read_bytes();mutants=[]
for name in ['fragment-id','undef-id','adjacent-judgment']:
    p=WORK/(name+'-run.stdout');candidate=p.read_bytes();position=next((i for i,(a,b) in enumerate(zip(truth,candidate)) if a!=b),min(len(truth),len(candidate)))
    assert truth!=candidate and not (WORK/(name+'-run.stderr')).read_bytes()
    status=json.loads((WORK/(name+'-run.json')).read_text());assert status['exit']==0
    mutants.append({'name':name,'first_different_byte':position,'exit':status['exit'],'stderr_bytes':0})
(DEST/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
owned=ROOT/'stage1/cohere/typeaware/rules/wave-22-sixth'
paths=[p for p in owned.rglob('*') if p.is_file()]+[ROOT/'bridge/tsgo/checker'/name for name in ['symbol_declaration_syntax.go','symbol_declaration_syntax_test.go','facts.go']]
(DEST/'source-hashes.json').write_text(json.dumps({str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in paths},indent=2)+'\n')
(DEST/'streams.json').write_text(json.dumps(records,indent=2)+'\n')
print('preserved',len(records),'streams/status files; mutants',json.dumps(mutants),flush=True)
