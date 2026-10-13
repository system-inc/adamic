from pathlib import Path
import json,subprocess,collections
out=Path('review/compiler/loom-slice-2-2')
def git(*args):return subprocess.check_output(['git',*args],text=True).strip()
def rows(text):return {s.split('|')[1].strip():s for s in text.splitlines() if s.startswith('| ') and s.split('|')[1].strip().endswith(('.a','.ts'))}
base=git('rev-parse','origin/main');old=rows(git('show',base+':internal/oracle/counts.md'));now=rows(Path('internal/oracle/counts.md').read_text())
owner={r['path']:r['owner'] for r in json.loads(Path('review/compiler/chain-slice-2/counts-attribution.json').read_text())['rows']}
changes=[dict(path=p,owner=owner.get(p),before=old.get(p),after=now.get(p)) for p in sorted(old.keys()|now.keys()) if old.get(p)!=now.get(p)]
counts=dict(base=base,base_rows=len(old),after_rows=len(now),added=sum(r['before'] is None for r in changes),changed_existing=sum(r['before'] is not None and r['after'] is not None for r in changes),removed=sum(r['after'] is None for r in changes),rows=changes)
assert counts['added']==46 and counts['changed_existing']==counts['removed']==0
assert set(r['path'] for r in changes)==set(owner) and all(r['owner'] for r in changes)
(out/'counts-attribution.json').write_text(json.dumps(counts,indent=2)+'\n')
paths=git('diff','--name-only','origin/main','--','internal/native/runtime').splitlines()
assert paths==['internal/native/runtime/adamic.h','internal/native/runtime/exceptions.c','internal/native/runtime/object.c']
def runtime_patch(a,b=None):
 args=['diff','--unified=0',a]+([b] if b else [])+['--','internal/native/runtime']
 return '\n'.join(line for line in git(*args).splitlines() if not line.startswith(('index ','@@')))
assert runtime_patch('199174800998904d8d646dfa5892f512cd5cb20f','0e49e2bc')==runtime_patch('origin/main')
(out/'runtime-own-diff.patch').write_text(git('diff','origin/main','--','internal/native/runtime')+'\n')
times=[]
for name in ['oracle','fixtures','support','ir-support','reader-products','fresh-archives','readers']:
 for line in (out/('finish-'+name+'.jsonl')).read_text().splitlines():
  if not line.startswith('{'):continue
  e=json.loads(line)
  if e.get('Test') and e['Action']=='pass':times.append(dict(check=name,package=e['Package'],test=e['Test'],seconds=e['Elapsed']))
top=[r for r in times if '/' not in r['test']]
assert len([r for r in top if r['check']=='oracle'])==74
assert len([r for r in top if r['check'] in ['support','ir-support']])==74
assert len([r for r in top if r['check']=='readers'])==4
assert all(r['seconds']<60 for r in top)
(out/'finish-test-seconds.json').write_text(json.dumps(times,indent=2)+'\n')
new=[r for r in top if r['check'] in ['reader-products','fresh-archives','readers'] or r['test'].startswith('TestFixturesSelfCompare') or (r['check']=='oracle' and any(n in r['test'] for n in ['NumberEvery','NullableObjectMap','HeldUnion','BooleanFilter']))]
assert len(new)==27
(out/'new-leaf-seconds.json').write_text(json.dumps(new,indent=2)+'\n')
mutants=[]
for name in ['mutants-overlay-results.json','prior-callback-results.json','callback-mutants/results.json','mutants-self-compare-results.json','weak-isolated/result.json','registry-mutant/result.json','grain-mutants.json','callback-mutants/nullable-token-omitted/artifact-result.json']:
 data=json.loads((out/name).read_text());mutants.extend(data if isinstance(data,list) else [data])
assert all(r['caught'] for r in mutants)
checks=[json.loads(s) for s in (out/'finish-checks.jsonl').read_text().splitlines()]
assert any(r['name']=='counts' and r['exit']==0 for r in checks)
assert any(r['name']=='vet' and r['exit']==0 for r in checks)
assert any(r['name']=='readers' and r['exit']==0 for r in checks)
lines=[
 'Built: slice 2 re-cut on main; all six original over-budget leaves split, with every assertion retained.',
 'Commit SHAs: main '+base+'; source 0a757e0ce28a55fa82797cfc46f68211972e4cdd; one delivery commit, whose SHA is in the final response and git log -1.',
 'Commands and outputs: 74 oracle leaves, 74 support leaves, four readers, three self-compare fixtures, and all eight preparation cases pass; vet passes; counts add 46 owned rows.',
 'Mutants: all 17 originals verified across the prior and finish runs; the 17th now fails the Node comparison with a narrowed-union stop. New ownership and reader guard mutants also fail.',
 'Not covered: full package suites, full gate, WASI, and macOS. The original first-green-within-30-minutes target was missed during cold setup.',
 '',
 'The branch starts at the fetched main tip above. Main moved once during this finish; the delivery commit was rebased once. A final fetch confirmed that main had not moved again. The original squash had no textual conflicts. The rebase conflicted only in internal/oracle/counts.md, which was restored from main and regenerated once on this base. Other overlapping edits merged automatically, preserving both sides.',
 '',
 'The two explicit Weak lifetime counters use main\'s keyed countCase registry. Counts move from '+str(counts['base_rows'])+' to '+str(counts['after_rows'])+' rows: 46 additions, zero changed existing rows, zero removals. Owners: '+', '.join(f'{k}: {v}' for k,v in sorted(collections.Counter(r['owner'] for r in changes).items()))+'. Every row and its values are recorded in counts-attribution.json.',
 '',
 'The runtime diff is exactly slice 2\'s own Error message repair in adamic.h (+2), exceptions.c (+15/-2), and object.c (+2). Its zero-context patch matches the pre-rebase slice after normalizing hunk locations and blob IDs. Main\'s runtime edits are retained. runtime-own-diff.patch records the complete diff. No code was copied from cohere.',
 '',
 '## Splits and every new leaf\'s seconds', '',
 'Single-fixture oracle leaves are split by backend; sanitized cases retain leak checks. Self-compare is split into closure, array-fill and sweep fixtures. The fixture ownership guard still requires every manifest row exactly once, and rejects missing, duplicate, mixed or nonexistent owners. Reader checks scan production, internal tests and external tests separately for each of native, lower, fresh and flow, preserving metadata and stale-entry assertions.',
 '',
 'Export preparation is split again into key discovery and archive cases. Serial key discovery avoids four simultaneous listings; parallel archive cases retain only Go\'s __.PKGDEF member, which go/importer reads, using the original ar envelope. Test-variant import aliases are preserved. Main\'s normal precomputed input index remains supported, and every archive case prepares itself when selected alone. All final preparation cases include setup in the seconds below.',
 '', '| Test | Seconds |', '|---|---:|']
for r in new:lines.append(f"| {r['test']} | {r['seconds']:.2f} |")
lines+=['','Measured with GOMAXPROCS=4 and a four-CPU cgroup quota. Scripted batches additionally restrict CPU affinity to four CPUs. nproc reports 5. new-leaf-seconds.json contains all 27 replacement and preparation cases; finish-test-seconds.json contains every completed affected leaf. The edited fixture ownership guard also passes.','', '## Exact finish commands and results','', '| Phase | Exit | Wall seconds | Command |', '|---|---:|---:|---|']
for r in checks:lines.append(f"| {r['name']} | {r['exit']} | {r['seconds']} | `{ ' '.join(r['command']).replace('|',chr(92)+'|') }` |")
lines+=['','The two IR support checks that did not build in the combined support command were completed with `timeout 180 go test ./internal/ir -run ^Test(CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues)$ -json -count=1 -timeout=90s`. Fresh archive preparation was interrupted after about 33 seconds by the combined package\'s 90-second limit, after the other seven cases passed. It was completed alone with `timeout 180 go test ./internal/ir -run ^TestProduct_CallTargetReaderArchivesFresh$ -json -count=1 -timeout=90s`, passing in 26.32 seconds. Final TestCallTargetReaders and changed-package vet pass.','', 'All test output was written to files. Tests were focused selectors, not whole package suites or the full gate. Source Node, JavaScript, native release, ASan, UBSan and Linux leak checks retain each fixture\'s original contract. The Weak freed-target witnesses retain their explicit native/Node lifetime difference.','', '## Every mutant and its catcher','', '| Mutant | Observed catcher |', '|---|---|']
for r in mutants:
 catcher=', '.join(r.get('markers',[])) or ', '.join(r.get('catchers',[])) or r.get('catcher','') or r.get('marker','')
 if r['mutant']=='split-numeric-fold': catcher='TestSplitSelfCompareAgreesWithNode: split native stdout differs from Node'
 if r['mutant']=='static-inline-definition': catcher='TestSplitSelfCompareAgreesWithNode: clang undefined-inline diagnostic'
 if r.get('structural_only'):catcher+='; structural only, not counted as a semantic mutant'
 lines.append('| '+r['mutant']+' | '+catcher.replace('|','\\|')+' |')
lines+=['| constant self-equality | TestSelfCompareConstantMutant: successful mutated build and execution disagree with Node for NaN |','', 'The 16 earlier original semantic proofs were accepted on base 199174800998904d8d646dfa5892f512cd5cb20f. The constant self-equality check also ran again on this base. The new Every release and sanitized leaves restore the ninth original adaptation witness; its JavaScript leaf passes that particular native-representation mutant. Together with prior-callback-results.json, all nine original numeric fixtures are covered. The current nullable-token mutant completes the original set of 17.','', 'For nullable-token-omitted, the compiler artifact was built with the recorded lower/callback_widening.go overlay. Node observes stdout `1,null\\n`, empty stderr, exit 0. Both native artifacts observe the narrowed-union panic, exit 70. Release build took 7.083 seconds and sanitized build 10.030 seconds; both builds succeeded. ASan/UBSan were enabled for the latter. The byte comparison catches the mismatch; no sanitizer fault is claimed. Exact commands, stdout, stderr and timings are in callback-mutants/nullable-token-omitted/artifact-results.json. The first direct Go-test attempt exhausted its outer limit during cold preparation and has no completed verdict; it is not credited. The artifact script initially expected a UBSan fault, but the recorded bytes show a checked stop, so its assertion and report were corrected without repeating either program.','', 'The ownership guard parses physical source. Its first Go-overlay mutant therefore had no effect on its observation and is not credited. The real source name was then changed temporarily, the guard failed for the missing fixture owner, and the original file was restored in finally. The reader allowlist omission fails for an unapproved read. Both results are in grain-mutants.json; mutant sources are .go.txt.','', 'The earlier union-liveness child completed after its collection supervisor ended; the complete log records FAIL in 9.14 seconds, while its process exit code is unknown. Timeouts and build failures are never credited as semantic kills.','', '## Preparation and failed attempts','', 'Setup reached its 600-second limit in go list -deps -export -json ./.... Its generated environment was usable and was sourced in later shells. The exact timing lines follow.','', '```text']
lines.extend(s for s in (out/'setup.log').read_text().splitlines() if s.startswith('setup:'))
lines+=['```','', 'The prior six budget failures remain historical evidence in overdue-leaves.json. Their replacement checks pass. Additional full-archive preparations timed out; retaining only export members reduced copying, but combined key discovery and archive cases still took 62.62 to 67.23 seconds under parallel preparation. Those cases were split into the eight final preparation cases above; the overdue cases were not repeated. The final combined package alarm interrupted Fresh below its share, so only that unfinished case was completed separately.','', 'I changed reader declarations while the combined support command was compiling; its generated IR test main referenced the preceding names. The other 72 support leaves completed and passed. After declarations settled, the two IR support leaves passed separately, all readers passed, and vet passed. The first per-package reader run also exposed a missing import alias for flow external tests, which was corrected and is covered by the final readers.','', 'Runtime archives were prepared through native.RuntimeRequestFor, RuntimeRequest.Build with four workers, and main\'s declared buildcache inputs. Go export and fixture-hook preparation ran outside budgeted fixture cases. Preparation commands had hard limits and retained logs. logs.tar.gz bundles raw logs and JSONL evidence. The initial historical reports and scripts remain available under review/compiler/chain-slice-2; checks.jsonl and finish-checks.jsonl distinguish prior and finish attempts.','']
(out/'REPORT.md').write_text('\n'.join(lines))
print(json.dumps(dict(base=base,counts={k:v for k,v in counts.items() if k!='rows'},new_leaves=len(new),maximum_new_seconds=max(r['seconds'] for r in new)),indent=2))
