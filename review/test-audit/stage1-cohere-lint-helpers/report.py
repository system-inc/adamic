import pathlib,json,re,statistics,difflib,subprocess,shutil,collections
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-lint-helpers';names=(p/'rows.txt').read_text().splitlines();prod=[n for n in names if n!='TestHelperMutants'];runs=json.loads((p/'runs.json').read_text());sup=json.loads((p/'supplemental-runs.json').read_text());menu=json.loads((p/'menu.json').read_text());base=(p/'base.txt').read_text().strip();source=subprocess.check_output(['git','show',base+':stage1/cohere/lint/helpers/helpers_test.go'],cwd=r,text=True)
def events(log):
 out=[]
 for l in (p/log).read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out

def terminal(log):return {e['Test']:e['Action'] for e in events(log) if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ['pass','fail','skip']}
def error(log,row):
 for e in events(log):
  if e.get('Test','').split('/')[0]==row and e.get('OutputType')=='error':return e['Output'].strip()
 return None

def command(log):
 rr=next(x for x in runs+sup if x['log']==log);return 'ADAMIC_MUTANT='+rr['selector']+'; '+ ' '.join(rr['command'])

matrix=[]
for m in menu:
 log=m['id']+'-matrix.log';t=terminal(log);matrix.append(dict(id=m['id'],rows=prod,failed=[n for n in prod if t.get(n)=='fail'],results=t,unknown=[n for n in prod if n not in t],command=command(log),log=log,bounded=True));assert not matrix[-1]['unknown']
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
rows=[]
for n in names:
 vals=[]
 for i in range(1,4):
  for e in events(n+'-timing-'+str(i)+'.log'):
   match=re.search(r'\t([0-9.]+)s\n$',e.get('Output',''))
   if 'Test' not in e and match:vals.append(float(match[1]))
 assert len(vals)==3
 kills=[m['id'] for m in matrix if n in m['failed']];unique=[m['id'] for m in matrix if m['failed']==[n]];verdict='witness' if n=='TestHelperMutants' else 'sacred' if unique else 'subsumed';subs=[] if verdict!='subsumed' else ['TestHelpersMatchCohere']
 log='W1.log' if verdict=='witness' else kills[-1]+'-matrix.log';id='W1' if verdict=='witness' else kills[-1];failure=error(log,n);pk=['P'] if n in prod and terminal('P.log').get(n)=='fail' else [];match=re.search(r'^func '+n+r'\(',source,re.M);line=source[:match.start()].count('\n')+1
 oracle={
 'TestHelpersMatchCohere':('Go encoding/json, cohere optionschema, generated Go target decoders and policy.Messages.Render; byte comparison. 5923 Expected seeds also check archived ESLint acceptance/refusal labels. Checked [{}] for require-description against the ESLint 10.8.1/plugin 4.8.1 archive; no fresh ESLint execution.',['external-run','external-authority']),
 'TestHelperMutants':('Independent Go answers and archived ESLint seeds against compiled builtin mutants, using its own direct byte-equality check. W1 proves that local check; WShared+M1 shows this witness does not protect the production compare helper.',['external-run','external-authority']),
 'TestMessageRefusalsMatchGo':('Go policy.Messages.Render panic text; own adamic prefix and exit 70. Full stderr rejects M4 despite the same exit code. Supplemental SOut shows stdout is unchecked.',['external-run','self']),
 'TestKnownGapsAreExplicit':('Handwritten NotYet labels and valid recovery answer, checked on source Node and sanitized native. No outside authority for those labels.',['external-run','self'])}[n]
 rows.append(dict(test=n,package='stage1/cohere/lint/helpers',file='stage1/cohere/lint/helpers/helpers_test.go:'+str(line),seconds=statistics.median(vals),timing_samples=vals,oracle=oracle[0],oracle_kind=oracle[1],kills=kills,unique_kills=unique,last_proven_fail=id+': '+str(failure),verdict=verdict,subsumed_by=subs,mutants_in_matrix=[] if n=='TestHelperMutants' else [m['id'] for m in menu],probe_kills=pk,subsumer_seconds=None,vacuous=False if pk else None,bounded=True,matrix_rows=prod if n in prod else ['TestHelperMutants'],evidence=command(log)+'; '+str(failure)+'; log='+log,probe_evidence=None if n not in prod else dict(command=command('P.log'),failure=error('P.log',n),entry='main.ts',diff='P.diff')))
for row in rows:
 if row['verdict']=='subsumed':row['subsumer_seconds']=rows[0]['seconds'];row['limitations']='Subsumption rests on '+str(len(row['kills']))+' caught mutant. It is a bounded hint, not deletion advice.'
(p/'rows.json').write_text(json.dumps(rows,indent=2))
# Remove a regex control-flow false positive; declarations and callback sites remain original-source locations.
f=p/'functions.json';inventory=json.loads(f.read_text());inventory['named']=[x for x in inventory['named'] if x['name']!='.if'];inventory['scope']='28 named functions/constructors including module entry, plus 11 callback expression sites. Declaration/caller inventory from source, not a dynamic branch coverage claim.';f.write_text(json.dumps(inventory,indent=2))
checks=[]
for log in ['W1.log','P.log','WShared-M1.log','SOut-matrix.log']:
 checks.append(dict(log=log,command=command(log),results=terminal(log),errors=[dict(test=e.get('Test'),line=e['Output'].strip()) for e in events(log) if e.get('OutputType')=='error']))
(p/'checks.json').write_text(json.dumps(checks,indent=2))
observations=[]
wanted=(p/'oracle-answers.log').read_text().splitlines()
for m in menu:
 got=(p/(m['id']+'-native-output.log')).read_text().splitlines();first=next((i for i,(a,b) in enumerate(zip(got,wanted),1) if a!=b),None);observations.append(dict(id=m['id'],line=first,got=got[first-1] if first else None,want=wanted[first-1] if first else None,command=command(m['id']+'-native-output.log'),log=m['id']+'-native-output.log'))
(p/'native-witnesses.json').write_text(json.dumps(observations,indent=2))
known=set();skipped=set()
for log in ['baseline.log','baseline-narrowed.log']:
 t=terminal(log);known.update(n for n in names if t.get(n)=='pass');skipped.update(n for n in names if t.get(n)=='skip')
assert known==set(names)
(p/'baseline-coverage.json').write_text(json.dumps(dict(passed=sorted(known),skipped=sorted(skipped),unknown=[]),indent=2))
# One copied authority value is checked against the archived external sample.
cases=json.loads(pathlib.Path('/tmp/u114/cases.json').read_text());sample=json.loads((r/'cohere/internal/lint/optionschema/testdata/samples.json').read_text());c=next(c for c in cases['Cases'] if c.get('Expected') and c['Rule']=='@eslint-community/eslint-comments/require-description' and c['Input']=='[{}]');s=next(s for s in sample['samples'] if s['rule']==c['Rule'] and s['elements']==[{}]);assert c['Expected']=='valid' and s['eslint']=='accepts'
(p/'authority-check.json').write_text(json.dumps(dict(case=c,archived_sample=s,authority_versions=sample['generatedFrom'],archive='cohere/internal/lint/optionschema/testdata/samples.json:4',cohere_pin='7945d102a6c18dd36adf9114a758ce646e8b2359',fresh_eslint_run=False),indent=2))
validations=[]
for f in sorted(p.glob('*.diff')):
 q=subprocess.run(['git','apply','--check','--cached',str(f)],cwd=r,capture_output=True,text=True);validations.append(dict(diff=f.name,exit=q.returncode,output=q.stdout+q.stderr));assert q.returncode==0
(p/'diff-validation.json').write_text(json.dumps(validations,indent=2))
for suffix in ['audit','follow','report']:
 shutil.copyfile('/tmp/u114-'+suffix+'.py',p/(suffix+'.py'))
metadata=dict(base=base,nproc=5,toolchain_setup_seconds=0,toolchain='Warm env.sh worked; setup skipped. API npm ci reported 428 ms. No other node_modules required.',binary_baseline_seconds=90.017,narrowed_baseline_seconds=35.179,native_build_wall={m['id']:next(x['wall'] for x in runs if x['log']==m['id']+'-native-build.log') for m in menu},compiler_build_wall=next(x['wall'] for x in runs if x['log']=='compiler-build.log'),timing_command_wall_sum=sum(x['wall'] or 0 for x in runs if '-timing-' in x['log']),timing_recovered='The first witness timing completed after its controller received SIGINT during menu correction. Its binary line is complete and retained; command wall is null.',audit_command_wall_sum=sum(x['wall'] or 0 for x in runs),supplemental_command_wall_sum=sum(x['wall'] for x in sup),cases=len(cases['Cases']),go_output_lines=len(wanted),elapsed_note='Approximately 20 minutes through evidence preparation. Timings were serial, each count=1; no concurrent test workload was launched.')
(p/'timing-summary.json').write_text(json.dumps(metadata,indent=2))
summary=['u114 at '+base[:12]+': all four listed tests exist; no families, helpers or vanished names.','Bounded verdicts: one sacred, two subsumed, one local-comparison witness.','All four production mutants are caught; each standalone diff builds natively.','All three production rows reject the empty-main probe; witness was not probed.','Production source restored; evidence on test-audit/stage1-cohere-lint-helpers.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
text+='| ID | Origin file:line | One-line change | Failed rows |\n|---|---|---|---|\n'
for m in menu:text+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+m['old']+'` → `'+m['new']+'` | '+', '.join(next(x['failed'] for x in matrix if x['id']==m['id']))+' |\n'
text+='\nSurvivors: none in the bounded production matrix. Native witnesses are in native-witnesses.json: M1 changes valid/invalid at line 15159; M2 emits unsupported schema keyword type at line 1; M3 rejects null at line 15171; M4 exits 70 with an erroneous phrase-count panic at line 23350.\n\n'
text+='Separate witness and supplemental observations:\n\n- W1, helpers_test.go:141: replace the witness equality comparison of got/want with got/got. All four builtin-mutant subcases fail at helpers_test.go:142, compiled mutant survived. W1 is not a production kill.\n- WShared, helpers_test.go:105: disable the production compare helper by comparing got/got. With admissible M1 also present, TestHelpersMatchCohere and TestHelperMutants both pass. This shows the witness does not protect the production helper. Its witness verdict applies only to its own direct byte-equality comparison.\n- SOut, main.ts:7: insert console.log of audit unexpected stdout. This insertion is outside the fixed menu, explicitly supplemental, and supports no verdict. Helpers agreement and gap labels fail; TestMessageRefusalsMatchGo passes. A native refusal run prints the unexpected line and still exits 70 with the expected stderr.\n\n'
notes=[
'The complete cold baseline exceeded 90 seconds during the final builtin-mutant subcase, after the main agreement row passed in 63.26 seconds. This was a budget timeout, not an observed assertion failure. The remaining three rows passed a narrowed clean run in 35.179 seconds.',
'The fetched origin/main is ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. All four names exist in go test -list; the package has no additional top-level Tests and no grouping is needed.',
'The production matrix is bounded to the three production rows. The builtin witness is excluded from production kill counts by the brief. Its guard is weakened separately; package-wide uniqueness is left for central replay.',
'The main agreement test compares source Node before building/running its native product. Mutant runs can stop at that first mismatch. Every standalone mutant was therefore separately built and executed natively against unchanged independent Go output; none is a compiler-warning kill.',
'The first planned fourth edit loosened equality to greater-than, which was not a literal fixed-menu operation. Before any mutant was applied, it was replaced by count+1, a strict off-by-one bound. The timing controller was interrupted, source cleanliness verified, and completed binary timing logs reused.',
'The witness has its own equality check and does not call production compare. Disabling compare while numeric equality is broken makes both the agreement row and witness pass. Formal witness status here is local to its direct equality check; it must not be interpreted as protection of the production comparator.',
'The refusal row checks exact stderr and code 70, so M4 is caught despite producing the expected exit code. It does not check stdout: the supplemental inserted output survives that row. This survivor is row-local and supplemental, not an unguarded whole-unit verdict.',
'The gap row expects handwritten NotYet labels and a handwritten valid recovery result. These are self expectations, not Go diagnostic authority. Source Node is an additional implementation execution. M2 changes the refusal reason and proves the label check can fail.',
'The Go oracle also checks 5923 pinned ESLint sample expectations. One valid [{}] require-description value was checked against the archived ESLint 10.8.1/plugin 4.8.1 sample. ESLint itself was not rerun or installed; these pins are external-authority evidence, separate from this session\'s Go execution.',
'The current gzip corpus has 23539 cases and 23609 output lines. Message values can contain newlines. Historical REPORT.md numbers of 22347 cases are stale and were not used as scope or timing evidence.',
'The declaration inventory lists 28 named functions/constructors including module entry and 11 callback sites. It is a source caller inventory, not dynamic branch coverage. A regex false positive for an if statement was removed from the inventory.',
'The first witness-edit draft matched both equality conditions and stopped before mutations. The supplemental comparator draft likewise needed a newline anchor. Both controller errors are preserved; they support no verdict.',
'The port selector is carried in an existing copied file, options_json.ts, so the witness\'s fixed copied-file list and the Go oracle remain unchanged. Switching /tmp/u114-mutant changes runtime behavior without changing source products. Standalone replay diffs contain no selector.',
'The empty-main probe is a separate native-buildable empty entry. It is not mixed into the production mutant switch or counted as a kill. It rejects all three production rows. No empty probe was attributed to the witness.',
'The subsumption of each smaller row rests on one caught mutant. Both smaller rows are faster than the 12.769-second subsumer. These are bounded hints, not advice to delete cheaper tests.',
'All timings use three separate count=1 invocations with a single top-level row and no concurrent audit workload. The first witness timing completed after its controller was interrupted; its binary line is valid, while command wall time is unknown.',
'No opt-ins or skips were found in this package. Warm tools did not excuse dependency setup: npm ci ran in stage3/api before the baseline. The Node loader needs no additional node_modules for this port.',
'No other repository packages were tested, no full integration gate or exhaustive input fuzzer was run, and no central repo-wide uniqueness replay was attempted. Oracle and compiler executables were built only as required for this unit.',
'Every standalone diff, including the empty probe, weakened checks and supplemental stdout insertion, applies to the starting commit and has its appropriate build or vet evidence. Production source is restored. No main push or pull request is made.'
]
text+='Brief ambiguities, mistakes, costs and limits:\n\n'+'\n'.join('- '+x for x in notes)+'\n\nSetup/build/run measurements:\n\n```json\n'+json.dumps(metadata,indent=2)+'\n```\n';(p/'REPORT.md').write_text(text)
print(json.dumps(metadata,indent=2));print([(x['test'],x['verdict'],x['kills']) for x in rows])
