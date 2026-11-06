#!/usr/bin/env python3
"""One bounded full-latent file group, both compilers and meter identity.
Run from report checkout. Commits and plainly pushes each successful piece.
"""
import hashlib,io,json,os,pathlib,re,signal,subprocess,sys,tarfile,time
report=pathlib.Path(__file__).resolve().parent
names=sys.argv[1:];assert names and len(set(names))==len(names)
stem=names[0].replace('/','_').removesuffix('.ts').replace('.','_') if len(names)==1 else 'group_'+hashlib.sha256('\n'.join(names).encode()).hexdigest()[:12]
piece=report/'pieces'/stem
if piece.with_suffix('.tar.gz').exists():raise SystemExit('already recorded')
scratch=pathlib.Path('/workspace/scratch/v3-pieces')/stem;scratch.mkdir(parents=True,exist_ok=True)
start=time.monotonic();deadline=start+23*60
records={n:{'file':n,'mode':'full selected-file group in complete tsc closure','sides':{}} for n in names}
for label in ('main','head'):
 tree=pathlib.Path(f'/workspace/v3-{label}-adapted-complete')
 env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE='\n'.join(str(tree/n) for n in names))
 env.pop('V3_MUTANT_SKIP_NODE',None)
 for variant in ('meter','count'):
  output=scratch/f'{label}-{variant}.jsonl';log=scratch/f'{label}-{variant}.log'
  with log.open('w') as stream:
   process=subprocess.Popen([f'/workspace/scratch/v3-{label}-{variant}-bin',str(tree/'src/tsc/tsc.ts'),str(output)],env=env,stdout=stream,stderr=stream,start_new_session=True)
   try:code=process.wait(timeout=max(1,deadline-time.monotonic()))
   except BaseException:
    os.killpg(process.pid,signal.SIGTERM)
    try:process.wait(timeout=5)
    except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
    raise
   assert code==0,(label,variant,code)
 raw=(scratch/f'{label}-count.jsonl').read_bytes();baseline=(scratch/f'{label}-meter.jsonl').read_bytes()
 stripped=re.sub(rb'"(?:ir_measured|view_elements|view_element_checks)":(?:true|false|[0-9]+),',b'',raw)
 assert stripped==baseline,(label,'meter byte identity')
 rows=list(map(json.loads,raw.splitlines()));assert len(rows)==len(names)+1,(label,len(rows))
 seen=set()
 for row in rows[1:]:
  name=str(pathlib.Path(row['file']).relative_to(tree));assert name in records and name not in seen;seen.add(name)
  units=row['units'];findings=row['findings'];blocked={f['unit'] for f in findings if f['phase']=='lowering'}
  free=[u for u in units if u['status']=='attempted' and u['where'] not in blocked]
  records[name]['sides'][label]={'view_elements':sum(u['view_elements'] for u in units),'view_element_checks':sum(u['view_element_checks'] for u in units),
   'units':len(units),'attempted':sum(u['status'] in ('attempted','panic') for u in units),'panicked_units':sum(u['status']=='panic' for u in units),'ir_observed':sum(u['ir_measured'] for u in units),
   'finding_free_attempted':len(free),'finding_free_view_elements':sum(u['view_elements'] for u in free),'finding_free_view_element_checks':sum(u['view_element_checks'] for u in free),
   'meter_byte_identical':True,'baseline_sha256':hashlib.sha256(baseline).hexdigest(),'source_sha256':hashlib.sha256((tree/name).read_bytes()).hexdigest()}
 assert seen==set(names)
 (scratch/f'{label}-pair.jsonl').write_bytes(baseline.replace(str(tree).encode(),b'/V3_TREE'))
for record in records.values():assert record['sides']['head']['source_sha256']==record['sides']['main']['source_sha256'],'different adapted bytes'
with (scratch/'pair.log').open('w') as stream:
 subprocess.run(['python3',str(report.parent/'pair/pair.py'),str(scratch/'main-pair.jsonl'),str(scratch/'head-pair.jsonl'),'/workspace/v3-main-adapted-complete',
 '--typescript',str(report.parents[1]/'api/node_modules/typescript'),'--json',str(scratch/'pair.json'),'--report',str(scratch/'pair.md')],stdout=stream,stderr=stream,check=True,timeout=max(1,deadline-time.monotonic()))
summary={'files':names,'records':list(records.values()),'elapsed_seconds':time.monotonic()-start}
(scratch/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
files={p.name:p.read_bytes() for p in scratch.iterdir() if p.is_file()}
files['SHA256SUMS']=''.join(f'{hashlib.sha256(data).hexdigest()}  {n}\n' for n,data in sorted(files.items())).encode()
piece.parent.mkdir(exist_ok=True)
with tarfile.open(piece.with_suffix('.tar.gz'),'w:gz') as archive:
 for n,data in sorted(files.items()):
  info=tarfile.TarInfo(n);info.size=len(data);info.mtime=0;archive.addfile(info,io.BytesIO(data))
piece.with_suffix('.json').write_text(json.dumps(summary,indent=2)+'\n')
subprocess.run(['python3',str(report/'update-report.py')],check=True)
with (scratch/'publication.log').open('w') as log:
 subprocess.run(['git','add',str(piece.with_suffix('.json')),str(piece.with_suffix('.tar.gz')),str(report/'REPORT.md'),str(report/'SHA256SUMS')],check=True,stdout=log,stderr=log)
 subprocess.run(['git','commit','-m','census: views-v3 price piece '+stem],check=True,stdout=log,stderr=log)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/views-v3-price'],check=True,stdout=log,stderr=log)
print(json.dumps(summary),flush=True)
