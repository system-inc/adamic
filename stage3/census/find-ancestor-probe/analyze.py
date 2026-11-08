import json,gzip,re,collections,html,hashlib,subprocess
from pathlib import Path
out=Path('/workspace/scratch/probe-report/stage3/census/find-ancestor-probe')
sites=json.loads((out/'sites.json').read_text())
root=Path('/workspace/scratch/fa-adapted/src')
# Positions in the TypeScript AST and checker locations use UTF-16 code units.
offset_cache={}
def pos(file,line,col):
 if file not in offset_cache:
  offsets=[0]
  for text in (root/file).read_bytes().decode('utf8').splitlines(keepends=True):offsets.append(offsets[-1]+len(text.encode('utf-16-le'))//2)
  offset_cache[file]=offsets
 return offset_cache[file][line-1]+col-1
results=[]
for side in ['parent','head']:
 selected=[];rows=[[] for s in sites]
 with gzip.open(out/(side+'-stderr.txt.gz'),'rt') as f:
  for number,line in enumerate(f,1):
   if not line.startswith('probe-binding\t'):continue
   parts=line.rstrip('\n').split('\t');assert len(parts)==6,(number,parts)
   m=re.fullmatch(r'/workspace/scratch/fa-adapted/src/(.*):(\d+):(\d+)',parts[1])
   if not m:continue
   for index,s in enumerate(sites):
    if m[1]!=s['file']:continue
    offset=pos(s['file'],int(m[2]),int(m[3]))
    if s['start']<=offset<s['end']:
     row=dict(stderr_line=number,line=line.rstrip('\n'));rows[index].append(row);selected.append(dict(site=s['site'],**row))
 with (out/(side+'-filtered-probes.txt')).open('w') as f:
  for row in selected:f.write(row['line']+'\n')
 (out/(side+'-filtered-index.json')).write_text(json.dumps(selected,indent=2)+'\n')
 results.append(rows)
comparisons=[]
for i,s in enumerate(sites):
 p,h=[collections.Counter(row['line'] for row in side[i]) for side in results]
 same=set(p)==set(h);both=bool(p) and bool(h)
 comparisons.append(dict(**s,parent=dict(p),head=dict(h),same=same,sequence_same=[r['line'] for r in results[0][i]]==[r['line'] for r in results[1][i]],observed_on_both=both,verdict='same' if same else 'different'))
(out/'binding-comparison.json').write_text(json.dumps(comparisons,indent=2)+'\n')
validations={s:json.loads((out/(s+'-stream-validation.json')).read_text()) for s in ['parent','head']}
binding_sites=[s for s in comparisons if s['observed_on_both']]
different=[s for s in binding_sites if not s['same']]
missing=[s for s in comparisons if not s['observed_on_both']]
parts=['# Probe binding census','', 'Task 448a51k. Parent `555b1645603b90aad7b6e6df25bc642abcc52708` logs `1060360b`; head `8afae94406430e5c3340b2b33d2657615b36cf72` logs `7216e359`. Report only.','',
 '## Method','',
 'Read c6f6688c\'s report and both findings-only lists first. One scratch checkout of 1060360b received both complete adaptation directories from eeb87728, then `stage3/apply.sh` ran once against TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Every adapted source hash matches c6f6688c\'s source-hashes.json. See [carry validation](carry-validation.json), [source validation](source-validation.json), [source hashes](source-hashes.json), and [apply log](apply.log).','',
 'Both guarded entry overlays and binaries were built from their own probe pin, using the unchanged shared checker. The head change from 63c11488 to 7216e359 touches only a test fixture. See [provenance](provenance.json) and [overlays](overlays/). Entry first-error runs use `LATENT_ASSERT_NO_OUTPUT=1`, `GOMEMLIMIT=5GiB`, `GOGC=70`, `ADAMIC_PROBE_BINDINGS=1`, and separate `LATENT_TIMES` files. Neither `LATENT_FULL` nor `LATENT_ONLY_FILE` is set. Each run uses src/tsc/tsc.ts and both no-output guards. All stderr is retained whole in [parent](parent-stderr.txt.gz) and [head](head-stderr.txt.gz). Wrappers record the child\'s own exit, stop at 1800 seconds without output or unit progress, and clean up process groups.','',
 'The supplied lists actually contain 30 parent array refusal-scan sites and five head method-bivariance sites, not 28 plus five. All 35 listed refusal-scan sites are covered, plus three controls. Controls are the first unchanged append call in lexical order in each of es2015.ts, es2018.ts and module/module.ts, excluding changed call spans and requiring identical original findings inside the control span. Two hook assignment sites have no enclosing call expression. Their bounded assignment region is retained as an absence check, not as evidence of equal resolved signatures.','',
 'The stock TypeScript 6.0.3 AST selects the innermost call containing each finding coordinate. Selection includes every probe line whose `where` lies from that call\'s start through its exclusive end, including nested calls. [Site spans and source](sites.json), [extractor](sites.cjs), [analysis](analyze.py). For each site, the table below quotes each distinct full line verbatim, with occurrence counts; every occurrence remains in the filtered files in stderr order. [Parent filtered lines](parent-filtered-probes.txt), [head filtered lines](head-filtered-probes.txt), and their [parent](parent-filtered-index.json) and [head](head-filtered-index.json) stderr line indexes retain duplicates. Same/different compares exact sets of full binding lines; repetition counts are shown separately and do not change that verdict.','',
 '## Finding stream validation','', '| Side | Child exit | Findings | Prior findings | Identical findings | Entire JSONL identical |','| --- | ---: | ---: | ---: | --- | --- |']
for side,v in validations.items():
 status=json.loads((out/(side+'-entry-run.status.json')).read_text())
 parts.append(f'| {side} | {status["exit"]} | {v["actual_findings"]} | {v["expected_findings"]} | {v["identical"]} | {v["raw_stream_identical"]} |')
parts+=['','Finding comparison preserves order, multiplicity, and every field of every finding, including phase, source coordinate, reason and text. Validation compares parent against c6f6688c parent-entry-run.jsonl and head against c6f6688c head-entry-run.jsonl without path normalization. See [parent validation](parent-stream-validation.json), [head validation](head-stream-validation.json), and the raw parent-entry-run.jsonl and head-entry-run.jsonl streams.','', '## Site comparison','']
for i,s in enumerate(comparisons,1):
 parts += [f'### {i}. {s["site"]} ({s["category"]})','', '```typescript',s['source'].replace('\r\n','\n'),'```','',f'Range: `{s["start_where"]}` to `{s["end_where"]}` (exclusive). '+('Enclosing call: `'+s['call']+'`.' if s['call'] else 'No enclosing call expression; assignment-region absence check.'),'', '| Parent | Head |','| --- | --- |']
 def cell(rows):
  if not rows:return 'No probe-binding lines.'
  return '<br>'.join('<strong>'+str(n)+' occurrence'+('s' if n!=1 else '')+'</strong><pre>'+html.escape(line).replace('|','&#124;')+'</pre>' for line,n in rows.items())
 parts += ['| '+cell(s['parent'])+' | '+cell(s['head'])+' |','', '**Verdict: '+s['verdict']+'.**'+(' No binding observed on both sides; this is not a resolved-signature equality verdict.' if not s['observed_on_both'] else ' Resolved signature, argument and constraint compared as quoted above. Ordered occurrence sequences are '+('identical.' if s['sequence_same'] else 'different; see the indexed filtered evidence.')),'']
parts+=['## Interpretation','',f'{len(binding_sites)} sites have bindings on both sides; {len(different)} of those differ. {len(missing)} sites lack binding observations on at least one side. This paragraph is provisional until the quoted differences are reviewed.','', '## Reproduction and checkpoints','', 'Preparation and measurement commands, exact flags, own child exits, durations and silence stops are retained in the status and pid JSON files. The setup log is setup.log. Controllers are [prepare.py](prepare.py) and [measure.py](measure.py). Separate checkpoint refs were pushed after preparation and each run so the final branch can have one directory-only commit directly on 1060360b without force-pushing. No main ref is pushed.','']
(out/'REPORT.md').write_text('\n'.join(parts))
print(json.dumps(dict(sites=len(sites),observed_both=len(binding_sites),different=len(different),missing=[s['site'] for s in missing]),indent=2))
