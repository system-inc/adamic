import pathlib,json,subprocess,os,time,hashlib,resource,datetime,socket
out=pathlib.Path(__file__).resolve().parent
selected=json.loads((out/'selection.json').read_text()); adapted=pathlib.Path('/tmp/adamic-gate/stage3-meter.EyQEeg/main-adapted'); names='|'.join(x['name'] for x in selected)
metadata={'main_sha':'161b02663a129f2f7da97fc08fc925b120302612','merge_sha':'b5733921511af8146a299502b31407dec3a1ab74','hostname':socket.gethostname(),'run_order':['main','merge'],'cap_seconds_per_side':600,'latent_mode':'first-error','selection_order':'source order in utilitiesPublic.ts','shared_adapted_tree':str(adapted),'timer_scope':'Existing latent UnitTimer: registration, lowering and panic recovery; excludes Load, refusal scan and timer record writing. Full entry graph retained. Unselected file and statement attempts skipped by scratch overlay. Facts and checker caches are cold at process start and shared among selected attempts as in latentRun.','binaries':{},'adapted_sha256':{},'runs':{}}
for p in sorted((adapted/'src').rglob('*')):
 if p.is_file():metadata['adapted_sha256'][str(p.relative_to(adapted))]=hashlib.sha256(p.read_bytes()).hexdigest()
for side in ['main','merge']:
 binary=out/(side+'-census');metadata['binaries'][side]=hashlib.sha256(binary.read_bytes()).hexdigest(); env=os.environ.copy();env.update(LATENT_SELECT_NAMES=names,LATENT_TIMES=str(out/(side+'-times.jsonl')),LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1'); started=time.monotonic(); stamp=datetime.datetime.now(datetime.timezone.utc).isoformat()
 with (out/(side+'-run.log')).open('w') as log:
  run=subprocess.run(['timeout','--kill-after=10s','600',str(binary),str(adapted/'src/tsc/tsc.ts'),str(out/(side+'-entry.jsonl'))],env=env,stdout=log,stderr=subprocess.STDOUT)
 elapsed=time.monotonic()-started;metadata['runs'][side]={'started_utc':stamp,'process_wall_seconds':elapsed,'exit_code':run.returncode};(out/'metadata.json').write_text(json.dumps(metadata,indent=2)+'\n');print(side,run.returncode,elapsed,flush=True)
 rows=[json.loads(x) for x in (out/(side+'-times.jsonl')).read_text().splitlines()]; actual={x['where'].split('/main-adapted/')[-1] for x in rows}; expected={x['where'] for x in selected}
 if run.returncode!=0 or len(rows)!=23 or actual!=expected:raise SystemExit(f'{side}: incomplete selection; stop before next side')
 print(side,'all 23 units timed',flush=True)
