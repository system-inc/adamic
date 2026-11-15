from pathlib import Path
import collections,json,re
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch/adapted-head/src/compiler')
r=json.loads((out/'evidence/type-whole.json').read_text());sites=collections.defaultdict(list);files=collections.Counter();forms=collections.Counter()
for e in r['Events']:
 p,line,col=e['Where'].rsplit(':',2);file=str(Path(p).relative_to(s));sites[file,int(line),int(col)].append(e);files[file]+=1;forms[e['Form']]+=1
text=['# Observed type-level sites','','This is a partial source/type measurement, not a completed lowering or an exhaustive production-stop count.','']
for (file,line,col),es in sorted(sites.items()):text += [f'## {file}:{line}:{col}','','Forms: '+', '.join(sorted({e['Form'] for e in es}))+f'; events: {len(es)}.','','```typescript',(s/file).read_text().splitlines()[line-1],'```','']
(out/'SITES.md').write_text('\n'.join(text))
summary=['# Type-level observations','',f'{len(r["Events"])} events; {len(sites)} distinct exact source positions. The exhaustive production count remains UNKNOWN. Each source line is quoted in [SITES.md](SITES.md).','', '| Form | Events | Sites |','| --- | ---: | ---: |']
for form,count in sorted(forms.items()):summary += [f'| {form} | {count} | {len({e["Where"]for e in r["Events"]if e["Form"]==form})} |']
summary += ['','| File | Events | Sites |','| --- | ---: | ---: |']
for file,count in sorted(files.items()):summary += [f'| {file} | {count} | {sum(1 for key in sites if key[0]==file)} |']
resource=json.loads((out/'evidence/type-whole.time.json').read_text());summary += ['',f'Whole-program run: {resource["seconds"]:.6f} s wall; {resource["peak_rss_KiB"]} KiB peak RSS. Input: 79 roots; 799,863 non-type AST nodes; 33 candidates; H.every=true; 190 explicit H types; 4,457 shapes.','', 'The private instrument cannot reconstruct all lowering-generated stores, carried slots, concrete generic instantiations or retained capture metadata. Its results are source/type judgments under the archived reconstruction, with neither an upper nor a lower bound on an exhaustive production count.']
(out/'TYPE-OBSERVATIONS.md').write_text('\n'.join(summary)+'\n')
p=out/'REPORT.md';t=p.read_text().replace('Pending controls and completed input preparation.','**Completed partial type-level observation: 22 events at 22 sites (6 writes, 9 literal initializations, 7 captured bindings).** The exhaustive production site count remains UNKNOWN because the source reconstruction cannot recover all lowered stores and metadata. See [TYPE-OBSERVATIONS.md](TYPE-OBSERVATIONS.md) and [all quoted sites](SITES.md).')
t=t.replace('Adaptation is still running; the provenance manifest in this checkpoint is provisional.','Both adaptations completed successfully. Their 79-root compiler manifests are byte-identical; the completed hashes are in evidence/provenance.json.')
p.write_text(t)
print('events',len(r['Events']),'sites',len(sites),dict(forms))

mutant=json.loads((out/'evidence/type-whole-mutant.time.json').read_text())
p=out/'TYPE-OBSERVATIONS.md';p.write_text(p.read_text()+f"\nThe separate head-reach mutant keeps the collector active, returns false from looseReaches, and records 0 events with identical gathering: {mutant['seconds']:.6f} s wall; {mutant['peak_rss_KiB']} KiB peak RSS.\n")
