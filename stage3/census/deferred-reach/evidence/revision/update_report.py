"""Apply the mapped-binder diagnosis to the retained same-input reach ledger."""
import argparse,json,re,subprocess,gzip
from pathlib import Path

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',type=Path,default=Path('/workspace/adamic'))
p.add_argument('--tool-sha',default='a6e9df28d39e7850a7689fbe615a145697ffda84')
a=p.parse_args();root=a.root;out=root/'stage3/census/deferred-reach'
old='f91c61aa8a39585188c3b199ca1a56aa70ed55f5'
def saved(name):
 cache=out/'evidence/revision'/('previous-'+name+'.gz')
 if cache.exists():
  print('retained baseline artifact: '+name,flush=True)
  return gzip.decompress(cache.read_bytes()).decode()
 child=subprocess.run(['git','show',old+':stage3/census/deferred-reach/'+name],cwd=root,text=True,capture_output=True)
 print('git show '+name+' child exit: '+str(child.returncode),flush=True)
 child.check_returncode();return child.stdout
r=json.loads(saved('REPORT.json'));text=saved('REPORT.md')
source=root/'stage3/census/latent/evidence/deferred-reach/revision/mapping-diagnosis.json'
diagnoses={d['target']:d for d in json.loads(source.read_text())};rel=lambda s:'src/'+s.split('/src/',1)[1] if '/src/' in s else s
signature='signature only, no body';sixth='reached, stops at its signature';hidden="unreached hidden behind a census caller's earlier stop"
for u in r['units']:
 if u['where'] not in diagnoses:continue
 d=diagnoses[u['where']];assert d['declared_binders']==1 and d['recovered_binders']==0 and not d['concrete'] and d['first_stop_matches_baseline']
 u['mapping_diagnosis']=d
 if u['where']=='src/compiler/moduleNameResolver.ts:3414:1':
  assert not any(t['stage']=='lowerBody_entry' for t in d['trace'])
  u['classification']=sixth;f=d['first_stop'];u['signature_stop']={k:rel(v) if k in ['unit','where'] else v for k,v in f.items()};u['signature_stop']['call_site']=d['call_site']
classes=['reached',signature,'unreached with no call anywhere','unreached, called only from outside the census',hidden,sixth,None]
def counts(units):return {c or 'unclassified: request without concrete witness':sum(u['classification']==c for u in units) for c in classes}
core=[u for u in r['units'] if u['where'].startswith('src/compiler/core.ts:')]
r['counts']={'core':counts(core),'all':counts(r['units'])}
for file,entry in r['per_file'].items():entry['classes']=counts([u for u in r['units'] if u['where'].rsplit(':',2)[0]==file])
r['tooling_commit']=a.tool_sha
r['status']='502 of 504 classified; two unread-binder body entries retained unresolved without changing inference'
r['revision']={'previous_report':old,'recording_rows_moved_unreached_to_reached':0,'previously_unclassified_moved_to_signature_stop':1,'compiled_observation_code_unchanged':True,'diagnosis':'one declared binder and zero recovered mappings in each of the three requests','pair_reader_and_mutant_in_tooling_commit':True}
(out/'REPORT.json').write_text(json.dumps(r,indent=2)+'\n')
start=text.index('The tooling is complete.');end=text.index('\n\n## Scope and denominators',start)
text=text[:start]+'''The tooling is complete, including the paired reader and its removal mutant in the same commit. Core.ts is fully classified; 502 of 504 declarations fit the six requested classes. toSearchResult is reached and stops at its signature. The two remaining body entries have an unread binder because lowering cannot recover T through their record types. They remain unresolved at that compiler inference boundary. No recording repair was needed; zero rows moved from unreached to reached through a recording change.'''+text[end:]
text=text.replace('8fea9f1018dfb124b0c77b4bac702e06f968680f',a.tool_sha)
start=text.index('| Class |');end=text.index('\n\nReached rows',start)
lines=['| Class | core.ts / 207 | core.ts bodies / 123 | All / 504 | All bodies / 367 |','| --- | ---: | ---: | ---: | ---: |']
for c in classes:
 n=r['counts']['core'][c or 'unclassified: request without concrete witness'];m=r['counts']['all'][c or 'unclassified: request without concrete witness'];label=c or 'Unresolved: unread-binder body entry'
 lines.append(f'| {label} | {n} | {"excluded" if c==signature else n} | {m} | {"excluded" if c==signature else m} |')
text=text[:start]+'\n'.join(lines)+text[end:]
text=text.replace('Reached rows use the existing concrete-context flag. The body-entry exception is kept separate, so a false flag is not asserted to prove that no body was entered.', 'The reached body class uses the existing concrete-context flag. The sixth class records a reached request that stops before body entry. The two unread-binder body entries are kept unresolved, so a false flag is not asserted to prove zero body entry.')
text=text.replace('That repair is in this report commit so the tooling commit remains confined to stage3/census/latent.', 'The reader repair and its positive and removal-mutant tests are in the same tooling commit as the stream field.')
text=text.replace('the pair suite passed all 19 tests.', 'the revised pair suite passed all 20 tests, including the removal mutant turning the unchanged positive eligibility test red.')
start=text.index('## Unclassified observations');end=text.index('## Full-mode scope',start)
diagnosis='''## Mapped-binder diagnosis and signature stop

All three requests have one declared binder T and zero mappings recovered by lowering. A diagnostic rerun of only their three calling units reproduces the original first stops exactly. The causes and complete traces are retained in [the tooling cases](../latent/evidence/deferred-reach/revision/CASES.md) and each REPORT.json row's mapping_diagnosis.

- **createComputedCompilerOptions**, utilities.ts:9031:1: caller `utilities.ts:9042:1`, call `utilities.ts:9042:26`. T occurs behind a mapped record, keyof and indexed accesses. inferTypes does not walk those mapped record properties, so concreteTypes has no T. Its body enters and completes because it simply returns options. The false flag is caused by missing argument readback, not initializer ownership. Inference and the all-binders-mapped requirement are unchanged; this row remains unresolved at that boundary.
- **updateCreateProgram**, executeCommandLine.ts:981:1: caller `executeCommandLine.ts:968:1`, call `executeCommandLine.ts:974:5`. T is inside the anonymous host record's `createProgram: CreateProgram<T>` property. inferTypes does not inspect that record field and recovers no T. Its body enters with the binder unread, then stops at `executeCommandLine.ts:983:26`: **NotYet, a function returning T**. Recording is not success-only: the hook runs before the body, but the required fully mapped context is absent. The README's positive definition is correct; its false-implies-unexamined claim was corrected. This row remains unresolved.
- **toSearchResult**, moduleNameResolver.ts:3414:1: caller `moduleNameResolver.ts:2656:1`, call `moduleNameResolver.ts:2661:16`, passing undefined. An absent argument supplies no useful binder witness, and the return alias puts T under a record property. It stops in signature lowering at `moduleNameResolver.ts:3414:28`: **NotYet, a value of type T | undefined**. No lowerBody entry occurs. Its class is **reached, stops at its signature**; instantiated remains false.

The expanded audit_deferred fixture proves that a generic called only by a top-level initializer records true and its exact initializer witness. A second generic records true and its witness before its body fails. Both pass in legacy, full and speculative modes against the existing observation binary. An initializer-only dropped-witness mutant exits 1 at the new initializer assertion. Thus neither unit kind nor success-only recording explains these rows.

**Movement:** zero boolean changes, zero rows moved from unreached to reached by fixed recording, and one formerly unresolved row moved into the requested signature-stop class. The two other rows stop at the diagnosed inference limitation without an invented class or an unrequested inference change.

Only the three diagnostic callers, the extended fixtures and their old-binary comparisons, and the pair suite were rerun. No production or observation Go code changed. The existing whole-statecopy, 79/81-file byte-identity and meter-identity results remain applicable. No new full-mode run was needed.

'''
text=text[:start]+diagnosis+text[end:]
start=text.index('| File | Deferred |');end=text.index('\n\nZero deferred declarations',start)
lines=['| File | Deferred | Signature only | Bodies | Reached | No reference | Outside only | Hidden | Signature stop | Unresolved |','| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |']
for file,s in r['per_file'].items():
 c=s['classes'];values=[s['deferred'],c[signature],s['body_only'],c['reached'],c['unreached with no call anywhere'],c['unreached, called only from outside the census'],c[hidden],c[sixth],c['unclassified: request without concrete witness']]
 lines.append('| '+file.removeprefix('src/compiler/')+' | '+' | '.join(map(str,values))+' |')
text=text[:start]+'\n'.join(lines)+text[end:]
text=text.replace('`Pending` means the explicitly unclassified request cases.', '`Unresolved` means the two body entries whose binder mapping cannot be read back.')
text += '\nThe revision is reproducible with `python stage3/census/deferred-reach/evidence/revision/update_report.py --root /workspace/adamic`, followed by `python stage3/census/deferred-reach/audit_report.py`. The original generator in steps2-3 is retained as the previous checkpoint producer.\n'
(out/'REPORT.md').write_text(text)
print('updated 504 rows: signature stop=1; unresolved=2; recording movement=0')
