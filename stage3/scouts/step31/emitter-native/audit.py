#!/usr/bin/env python3
"""Recount delivery claims from captured processes and pinned source bytes."""
import argparse,gzip,hashlib,json,re,sys
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('tree',type=Path);parser.add_argument('--mutant',choices=['stop','native-status','source-hash','node-output','witness-output']);args=parser.parse_args()
root=Path(__file__).resolve().parent;e=root/'evidence'
summary=json.loads((e/'summary.json').read_text());stops=json.loads((e/'stops.json').read_text());hashes=json.loads((e/'source-hashes.json').read_text());witnesses=json.loads((e/'witness-results.json').read_text());specs=json.loads((root/'witnesses.json').read_text())
if args.mutant=='stop':stops[0]['message']='invented stop'
if args.mutant=='native-status':summary['native_executed']=True
if args.mutant=='source-hash':hashes[next(iter(hashes))]='0'*64
if args.mutant=='node-output':summary['node_output_sha256']='0'*64
if args.mutant=='witness-output':witnesses[0]['node']['stdout']='wrong\n'
assert len(stops)==summary['stops']==15,'stop population'
for row in stops:
    for split in [0,1]:
        prefix=e/f"stop-{row['order']:02}-split-{split}"
        assert prefix.with_suffix('.exit').read_text()=='1\n','captured build exit'
        diagnostic=gzip.decompress(prefix.with_suffix('.stderr.gz').read_bytes()).decode()
        first=re.search(r'^(?:adamic: )?(.+?):(\d+):(\d+): (.+)$',diagnostic,re.M)
        assert first and first.group(4)==row['message'],'first-stop recount'
        assert (int(first.group(2)),int(first.group(3)))==(row['line'],row['column']),'stop coordinate'
    original=row['original'];text=(args.tree/original['file']).read_text().splitlines()
    assert text[original['line']-1]==original['lineText'],'upstream source-line attribution'
    assert row['owner'] and (root/row['witness']).exists(),'witness and owner'
assert summary['native_attempts']==32, 'native attempt count'
for split in [0,1]:
    assert (e/f'native-{split}.exit').read_text()=='1\n','untouched native build exit'
    diagnostic=gzip.decompress((e/f'native-{split}.stderr.gz').read_bytes()).decode()
    first=re.search(r'^(?:adamic: )?(.+?):(\d+):(\d+): (.+)$',diagnostic,re.M)
    assert first and first.group(4)==stops[0]['message'],'untouched native first stop'
assert summary['native_executed'] is False and summary['native_comparison'] is None,'no native success from failed builds'
for name,wanted in hashes.items():assert hashlib.sha256((args.tree/name).read_bytes()).hexdigest()==wanted,'pinned adapted source hash: '+name
assert hashlib.sha256((e/'golden.jsonl').read_bytes()).hexdigest()==summary['node_output_sha256'],'stock Node output hash'
assert json.loads((e/'node-comparison.json').read_text())['success'],'Node comparison'
assert len(specs)==len(witnesses)==15,'witness population'
main=json.loads((e/'main-witness-results.json').read_text())
for spec,candidate,baseline in zip(specs,witnesses,main):
    assert candidate['name']==baseline['name']==spec['name'],'witness order'
    assert candidate['node']==baseline['node']==dict(exit=0,stdout=spec['stdout'],stderr=''),'witness Node output'
    source=(root/'witnesses'/(spec['name']+'.a')).read_text()
    assert source.splitlines()[0]=='// a-check: type error TS'+spec['checkerCode'],'repository a-check header'
    for observed in [candidate,baseline]:
        for build in observed['builds']:
            assert build['exit']==1 and 'error TS'+spec['checkerCode']+':' in build['diagnostic'],'real witness checker code'
        assert observed['mutant']['exit']==0 and observed['mutant']['stdout']!=spec['stdout'],'source mutant differs only after successful execution'
mutants=json.loads((e/'replay-mutant-results.json').read_text())
assert len(mutants)==12 and mutants[0]['report']['success'],'replay baseline and mutant population'
for row in mutants[1:]:assert row['comparison_exit']==1 and not row['report']['success'],'replay mutant caught'
print('PASS: 32 failed native attempts, 15 mapped stops/owners, both compiler witness records, source hashes, Node and mutant evidence')
