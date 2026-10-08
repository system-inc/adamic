#!/usr/bin/env python3
"""Collect raw scratch measurements; never copy a modified upstream source tree."""
import argparse,bisect,gzip,hashlib,json,subprocess
from pathlib import Path
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('annotations',type=Path)
args=parser.parse_args()
here=Path(__file__).resolve().parent
out=here/'evidence/step32'
annotations=json.loads(args.annotations.read_text())
def save(source,name):
    (out/(name+'.gz')).write_bytes(gzip.compress(Path(source).read_bytes(),mtime=0))
def run(stem,command):
    with (out/(stem+'.stdout.gz.tmp')).open('wb') as stdout,(out/(stem+'.stderr.gz.tmp')).open('wb') as stderr:
        result=subprocess.run(command,stdout=stdout,stderr=stderr)
    rawout=(out/(stem+'.stdout.gz.tmp'));rawerr=(out/(stem+'.stderr.gz.tmp'))
    stdout=rawout.read_text();stderr=rawerr.read_text()
    save(rawout,stem+'.stdout');save(rawerr,stem+'.stderr')
    rawout.unlink();rawerr.unlink()
    return result.returncode,stdout,stderr
comparison=json.loads(Path('/tmp/step32-snapshots/comparison.json').read_text())
(out/'comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
for file in Path('/tmp/step32-snapshots').iterdir():
    if file.suffix in ('.stdout','.stderr','.exit') or file.name=='binary.log': save(file,'snapshot-'+file.name)
for profile,prefix in [('current','/tmp/step32-final'),('reference','/tmp/step32-snapshots/build')]:
    for split in (0,1):
        for suffix in ('stdout','stderr'): save(f'{prefix}-{split}.{suffix}',f'{profile}-{split}.{suffix}')
save('/tmp/step32-snapshots/binary.log','binary.txt')
for suffix in ('stdout','stderr','exit'): save('/tmp/step32-node-version.'+suffix,'node-entry.'+suffix)
for file in ('apply','reference-apply','76','submodules','compiler-build','progress','snapshots'):
    save('/tmp/step32-'+file+'.log',file+'.txt')
save('/tmp/step32-reference-adapted/patch-set.md','reference-patch-set.md')
(out/'submodule-pins.txt').write_text(subprocess.check_output(['git','-C','/tmp/step32-compiler','submodule','status','--recursive'],text=True))
walk=json.loads(Path('/tmp/step32-progress/stops.json').read_text());maps={}
for row in walk:
    n=row['ordinal'];a=annotations[str(n)]
    row.update(a)
    file=row['file']
    if file not in maps:
        raw=(Path('/tmp/step32-adapted')/file).read_bytes().decode().encode('utf-16-le')
        maps[file]=[raw,list(range(len(raw)//2)),raw]
    current,mapping,original=maps[file]
    starts=[0]+[i+1 for i in range(len(current)//2) if current[2*i:2*i+2]==b'\n\x00']
    origin=mapping[starts[row['line']-1]+row['column']-1]
    original_starts=[0]+[i+1 for i in range(len(original)//2) if original[2*i:2*i+2]==b'\n\x00']
    line=bisect.bisect_right(original_starts,origin)-1
    row.update(original_line=line+1,original_column=origin-original_starts[line]+1)
    row['in_pristine']=row['message'] in Path('/tmp/step32-final-0.stderr').read_text()
    for split in (0,1):
        for suffix in ('stdout','stderr','exit'):save(f'/tmp/step32-progress/{n:02d}-split-{split}.{suffix}',f'walk-{n:02d}-{split}.{suffix}')
    replacement=Path(f'/tmp/step32-progress/{n:02d}-replacement.json')
    if replacement.exists():
        data=json.loads(replacement.read_text());start,end=data['start'],data['end']
        assert current[start*2:end*2].decode('utf-16-le')==data['original']
        new=data['replacement'].encode('utf-16-le')
        maps[file][0]=current[:start*2]+new+current[end*2:]
        mapping[start:end]=[mapping[start]]*(len(new)//2)
        (out/replacement.name).write_bytes(replacement.read_bytes())
    probe=here/row['probe']
    if n<=15:
        previous=comparison['stops'][n-1]
        assert row['message']==previous['old_message'], 'walk needs a distinct witness mapping'
        row.update(probe_types_exit=previous['probe_types_exit'],probe_build_exit=previous['probe_build_exit'],probe_first_stop=previous['probe_first_stop'])
    else:
        code,stdout,stderr=run(f'witness-{n:02d}-types',['/tmp/step32-adamic','types',str(probe)])
        row['probe_types_exit']=code
        code,stdout,stderr=run(f'witness-{n:02d}-build',['/tmp/step32-adamic','build',str(probe),'-o',f'/tmp/step32-probe-{n:02d}'])
        row.update(probe_build_exit=code,probe_first_stop=stderr.splitlines()[0] if stderr else None)
    code,stdout,stderr=run(f'witness-{n:02d}',['node',str(here/'node.mjs'),str(probe)])
    row.update(node_exit=code,node_stdout=stdout,node_stderr=stderr)
    mutant=Path(f'/tmp/step32-node-mutant-{n:02d}.a')
    mutant.write_text(probe.read_text()+'\nconsole.log("step32 changed computation");\n')
    code,stdout,stderr=run(f'mutant-{n:02d}',['node',str(here/'node.mjs'),str(mutant)])
    row.update(mutant_exit=code,mutant_stdout=stdout,mutant_stderr=stderr)
    assert code==0 and stderr=='' and stdout!=row['expected_stdout']
    print(f'{n}: source Node matches; semantic stdout mutant killed',flush=True)
(out/'walk.json').write_text(json.dumps(walk,indent=2)+'\n')
result={'compiler_commit':comparison['compiler_commit'],'skipped_merge':'8f32e51e8fc41b8f1177453213ca5453ce764486','combined_mode_measured':False,'publication_base':subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip(),'current_exits':[int(Path(f'/tmp/step32-final-{s}.exit').read_text()) for s in (0,1)],'reference_exits':[comparison['split_0_exit'],comparison['split_1_exit']],'reference_first_stop':comparison['first_stop'],'current_first_stop':Path('/tmp/step32-final-0.stderr').read_text().splitlines()[0],'current_source_hashes':{str(f.relative_to('/tmp/step32-adapted')):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(Path('/tmp/step32-adapted/src').rglob('*')) if f.is_file()},'binary_sha256':hashlib.sha256(Path('/tmp/step32-adamic').read_bytes()).hexdigest(),'nproc':int(subprocess.check_output(['nproc'],text=True))}
(out/'result.json').write_text(json.dumps(result,indent=2)+'\n')
