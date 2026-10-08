"""Collect train stop receipts, retaining pending rehearsal status."""
import bisect, hashlib, json, re, shutil
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
BUILD=Path('/tmp/tsc-train-build-sparse')
PROGRESS=Path('/tmp/tsc-train-progress')
WITNESSES=Path('/tmp/tsc-train-witnesses')
def digest(data):return hashlib.sha256(data).hexdigest()
def units(text):
    raw=text.encode('utf-16-le')
    return ''.join(chr(int.from_bytes(raw[i:i+2],'little')) for i in range(0,len(raw),2))
def read(path):return json.loads(path.read_text())
rows=read(PROGRESS/'stops.json');pristine=(BUILD/'build-0.stderr').read_text();maps={};replacements=[]
probes=sorted((ROOT/'stage3/drivers/tsc-entry/probes').glob('[0-9][0-9]-*.a'))
for row in rows:
    number=row['ordinal'];file=row['file']
    if file not in maps:
        original=units((BUILD/'adapted'/file).read_bytes().decode())
        maps[file]=[original,list(range(len(original))),original]
    current,mapping,original=maps[file]
    starts=[0]+[i+1 for i,c in enumerate(current) if c=='\n']
    offset=starts[row['line']-1]+row['column']-1
    origin=mapping[offset];original_starts=[0]+[i+1 for i,c in enumerate(original) if c=='\n']
    line=bisect.bisect_right(original_starts,origin)-1
    row.update(original_line=line+1,original_column=origin-original_starts[line]+1,phase='checker',status='pending',message_in_pristine=row['message'] in pristine)
    code=re.match(r'error (TS\d+):',row['message']).group(1);row['code']=code
    if number in (9,14):
        row['owner']={'workstream':'scratch continuation / inference','diagnostic_gate':'compiler checker','basis':'message absent from pristine diagnostics; throwing bodies can change inferred types','production_fix_claimed':False}
    else:
        family=('exact optional fields' if code in ('TS2375','TS2379') else 'checked indexed tuple reads' if code=='TS2488' else 'checked optional reads and arguments')
        row['owner']={'workstream':'compiler checked strictness','family':family,'candidate_branch':'codex/stricter-options-next','integration':'conflicted and aborted','diagnostic_gate':'internal/load/load.go:compilerOptions and load; embedded TypeScript checker','basis':'train uses Strict, NoUncheckedIndexedAccess and ExactOptionalPropertyTypes; no successful checked-options composition measured','production_fix_claimed':False}
    probe=probes[number-1] if number<=15 else HERE/'witnesses/16-symbol-array.a'
    stem=probe.stem;row['probe']=str(probe.relative_to(ROOT))
    for kind in ('node','native'):
        for suffix in ('stdout','stderr'):
            source=WITNESSES/(stem+'-'+kind+'.'+suffix) if number<=15 else Path('/tmp/tsc-train-16-'+kind+'.'+suffix)
            shutil.copyfile(source,HERE/'logs'/(stem+'-'+kind+'.'+suffix))
        exit_code=int((WITNESSES/(stem+'-'+kind+'.exit')).read_text()) if number<=15 else int(Path('/tmp/tsc-train-16-'+kind+'.exit').read_text())
        (HERE/'logs'/(stem+'-'+kind+'.exit')).write_text(str(exit_code)+'\n')
        row[kind+'_exit']=exit_code
    row['node_stdout']=(HERE/'logs'/(stem+'-node.stdout')).read_text()
    row['probe_message']=re.sub(r'^.+:\d+:\d+: ', '',(HERE/'logs'/(stem+'-native.stderr')).read_text().splitlines()[0])
    row['probe_exact_message_match']=row['probe_message']==row['message']
    for split in (0,1):
        for suffix in ('stdout','stderr','exit'):
            source=PROGRESS/f'{number:02d}-split-{split}.{suffix}';shutil.copyfile(source,HERE/'logs'/source.name)
    replacement=PROGRESS/f'{number:02d}-replacement.json'
    if replacement.exists():
        data=read(replacement);start,end=data['start'],data['end']
        assert units(data['original'])==current[start:end]
        replacement_units=units(data['replacement'])
        row['replaced_function']=data['function']
        replacements.append(dict(ordinal=number,file=file,function=data['function'],start=start,end=end,original_sha256=digest(data['original'].encode()),replacement=data['replacement']))
        maps[file][0]=current[:start]+replacement_units+current[end:]
        mapping[start:end]=[mapping[start]]*len(replacement_units)
(HERE/'stops.json').write_text(json.dumps(rows,indent=2)+'\n')
(HERE/'replacements.json').write_text(json.dumps(replacements,indent=2)+'\n')
for name in ['closure.json','outside-compiler.json']:
    shutil.copyfile(BUILD/name,HERE/name)
for split in (0,1):
    for suffix in ('stdout','stderr','exit'):shutil.copyfile(BUILD/f'build-{split}.{suffix}',HERE/'logs'/f'pristine-split-{split}.{suffix}')
shutil.copyfile(BUILD/'compiler-commit.log',HERE/'logs/compiler-commit.txt')
shutil.copyfile(BUILD/'adapted/patch-set.md',HERE/'patch-set.md')
shutil.copyfile(Path('/tmp/tsc-train-node.stdout'),HERE/'logs/node-entry.stdout')
shutil.copyfile(Path('/tmp/tsc-train-node.stderr'),HERE/'logs/node-entry.stderr')
# Pin every reached source, independently of subsequent disposable replacements.
closure=read(BUILD/'closure.json')
files=closure['files']
manifest={file:digest((BUILD/'adapted'/file).read_bytes()) for file in files}
(HERE/'source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(f'Collected {len(rows)} pending train stops and {len(replacements)} scratch-only replacements')
