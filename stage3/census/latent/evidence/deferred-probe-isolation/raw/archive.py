import collections,gzip,hashlib,json,pathlib,shutil,subprocess
R=pathlib.Path('/workspace/latent-deferred-probe-isolation');S=pathlib.Path('/workspace/ldpi')
E=R/'stage3/census/latent/evidence/deferred-probe-isolation';E.mkdir(parents=True,exist_ok=True)
A=S/'large-evidence';A.mkdir(exist_ok=True)
def sha(p):return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,indent=2,sort_keys=True)+'\n')
def zipped(src,dest):
 with dest.open('wb') as f:
  with gzip.GzipFile(fileobj=f,mode='wb',mtime=0,filename='') as z:z.write(src.read_bytes())
rows={k:[json.loads(l) for l in (S/f'combined-{k}/census.jsonl').read_text().splitlines()] for k in ('shallow','fixed')}
for k,v in rows.items():
 st=json.loads((S/f'combined-{k}/metadata.json').read_text());assert st['complete'] and all(x['child_exit_code']==0 and not x['deadline_expired'] for x in st['children'])
 assert v[0]['latent_mode']=='first-error' and not v[0]['checker_rejected'] and len(v[0]['sources'])==81
 assert v[-1]['status']=='relation_report'
 assert len([r for r in v if 'units' in r])==81
 zipped(S/f'combined-{k}/census.jsonl',A/f'{k}.jsonl.gz')
 shutil.copyfile(S/f'combined-{k}/metadata.json',A/f'{k}-metadata.json')
delta=json.loads((S/'delta.json').read_text())
checks={}
for kind in ('raw_findings','distinct_reads'):
 checks[kind]={x:len(delta[kind][x]) for x in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']}
 checks[kind]['totals']=delta[kind]['totals']
checks['streams_byte_identical']=(S/'combined-shallow/census.jsonl').read_bytes()==(S/'combined-fixed/census.jsonl').read_bytes()
checks['all_records_equal']=rows['shallow']==rows['fixed']
checks['files']=81
checks['units']={k:dict(collections.Counter(u['status'] for r in v for u in r.get('units',[]))) for k,v in rows.items()}
checks['finding_occurrences']={k:sum(len(r.get('findings',[])) for r in v) for k,v in rows.items()}
checks['relation_reports']={k:v[-1]['relation_overflow'] for k,v in rows.items()}
checks['runs']={k:[{f:x[f] for f in ['child_exit_code','deadline_expired','wall_seconds','peak_rss_kib']} for x in json.loads((S/f'combined-{k}/metadata.json').read_text())['children']] for k in rows}
for name in ['guard-fixed-final','guard-mutant-final','guard-boundary-mutant','latent-tests3','overlay-latent-tests','dependencies-final']:
 st=json.loads((S/name/'status.json').read_text());assert st['child_exit_code']==(1 if 'mutant' in name else 0)
 checks[name]={k:st[k] for k in ['child_exit_code','deadline_expired','wall_seconds','peak_rss_kib']}
 shutil.copyfile(S/name/'status.json',A/f'{name}-status.json')
 zipped(S/name/'output.log',A/f'{name}-output.log.gz')
 if 'guard-' in name:shutil.copyfile(S/name/'output.log',E/f'{name}.log.txt')
 else:
  if name=='dependencies-final':continue
  logrows=[json.loads(l) for l in (S/name/'output.log').read_text().splitlines()]
  assert not any(r['Action']=='fail' for r in logrows)
  checks[name]['passed_tests_and_subtests']=sum(r['Action']=='pass' and bool(r.get('Test')) for r in logrows)
  checks[name]['longest_test_seconds']=max(r.get('Elapsed',0) for r in logrows if r['Action']=='pass' and r.get('Test'))
files={str(pathlib.Path(p).relative_to(S/'adapted')):sha(p) for p in rows['shallow'][0]['sources']}
prov=dict(main=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip(),
 cohere=subprocess.check_output(['git','ls-tree','HEAD','cohere'],cwd=R,text=True).split()[2],
 typescript_source=subprocess.check_output(['git','rev-parse','HEAD'],cwd=S/'adapted',text=True).strip(),
 source_hashes=files,source_streams={k:sha(S/f'combined-{k}/census.jsonl') for k in rows},
 binaries={k:sha(S/f'{k}-split-census') for k in rows},
 artifact_branch='typescript/latent-deferred-probe-isolation-evidence')
fixed=json.loads((S/'fixed-split/overlay.json').read_text())['Replace'];shallow=json.loads((S/'shallow-split/overlay.json').read_text())['Replace']
diffs=[p for p in fixed if pathlib.Path(fixed[p]).read_bytes()!=pathlib.Path(shallow[p]).read_bytes()]
assert diffs==[str(R/'internal/lower/latent_hook.go')],diffs
assert pathlib.Path(fixed[diffs[0]]).read_text().replace('probe := *l.latentSnapshot()','probe := *l')==pathlib.Path(shallow[diffs[0]]).read_text()
prov['only_runtime_source_difference']='internal/lower/latent_hook.go: probe := *l versus probe := *l.latentSnapshot()'
prov['main_base']=prov.pop('main')
prov['go_version']=subprocess.check_output(['/workspace/adamic-tools/go/bin/go','version'],text=True).strip()
prov['stock_typescript_version']=json.loads(pathlib.Path('/workspace/latent-whole-passes/stage3/api/node_modules/typescript/package.json').read_text())['version']
prov['fix_source_sha256']={p:sha(R/p) for p in ['stage3/census/latent/lower.go.txt','stage3/census/latent/make_overlay.py','stage3/census/latent/deferred_probe_test.go.txt','stage3/census/latent/overlaytext/deferred_probe_test.go','cloud/fast-gate/compiler-dependencies.json']}
prov['scope']='stock first-error, full entry graph and registrations; disjoint complete-file observations retain every stock scan and unit body; no tail passes'
prov['reassembly_control_sha256']=sha(S/'fixture-control.py')

for k in ['fixed','shallow']:
 d=json.loads((S/f'{k}-split/overlay.json').read_text())['Replace']
 prov[k+'_runtime_overlay_sources']={str(pathlib.Path(p).relative_to(R)):sha(t) for p,t in d.items() if not p.endswith('_test.go')}
 shutil.copyfile(pathlib.Path(d[str(R/'internal/lower/latent_hook.go')]),A/f'{k}-hook.go.txt')
for source,target in [('stage3/census/latent/deferred_probe_test.go.txt','predicate-guard.go.txt'),('stage3/census/latent/overlaytext/deferred_probe_test.go','boundary-guard.go.txt'),('stage3/census/latent/lower.go.txt','fixed-template.go.txt')]:
 shutil.copyfile(R/source,A/target)
for item in ['run.py','split-overlay.py','fixture-control.py','combine.py','chunks.json','chunks.cjs','metadata-audit.py','metadata-delta.json','archive.py','summary.py','guard-files.json']:
 shutil.copyfile(S/item,A/item)
shutil.copytree(S/'slices',A/'slices',dirs_exist_ok=True)
# Archive bulky child streams/timers, retaining their actual status records.
for p in list((A/'slices').rglob('*')):
 if p.is_file() and p.suffix in ['.jsonl','.tsv','.log']:
  zipped(p,p.with_name(p.name+'.gz'));p.unlink()
shutil.copytree(S/'fixture-control',A/'fixture-control',dirs_exist_ok=True)
for p in (S/'split-fixture').glob('*.ts'):shutil.copyfile(p,A/(p.name+'.txt'))
fixture=json.loads((S/'fixture-control/summary.json').read_text());assert all(x['whole_vs_slices_file_records_equal'] for x in fixture.values())
checks['reassembly_control']=fixture
shutil.copyfile(S/'run.py',E/'run.py')
zipped(S/'delta.json',E/'delta.json.gz')
shutil.copyfile(S/'pair-report.md',A/'pair-report.md')
shutil.copyfile(S/'metadata-delta.json',E/'metadata-delta.json')
checks['metadata_delta']=json.loads((S/'metadata-delta.json').read_text())
dump(E/'checks.json',checks);dump(E/'provenance.json',prov)
for p in E.iterdir():
 if p.is_file():shutil.copyfile(p,A/p.name)
print(json.dumps(checks,indent=2))
