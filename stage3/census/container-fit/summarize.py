"""Render only observed ordinary-census results; latent answers stay explicitly unknown."""
import collections,gzip,hashlib,json,pathlib
O=pathlib.Path(__file__).resolve().parent;T='/workspace/container-fit-adapted/'
def rows(label):
 p=O/'evidence/per-file'/label/'all/records.jsonl.gz'
 if not p.exists():return []
 with gzip.open(p,'rt') as f:return [json.loads(x) for x in f]
def norm(x):return x.replace(T,'')
b,h=rows('base'),rows('head')
b={norm(r['roots'][0]):r for r in b if len(r['roots'])==1 and r['roots'][0].endswith('.ts')}
h={norm(r['roots'][0]):r for r in h if len(r['roots'])==1 and r['roots'][0].endswith('.ts')}
metrics={}
for label in ['base','head']:
 p=O/'evidence/per-file'/label/'all/metrics.json'
 if p.exists():metrics[label]=json.loads(p.read_text())
lines=['# Container fit census (zrnkpdz): latent measurement stopped','',
'The latent comparison is unavailable. The pinned measurement overlay fails before a head binary can be built: `lowering.containerFits: state copy: mutable named container containerFitPair requires an explicit copier`. No latent census process was started. No empty regression list, zero panic count, cap clearance, or fit improvement is claimed.','',
'## Pins and method','',
'Base: `d4c6b07aca1a7c55785226563c2ad79d552c1adf`. Head A: `1197ee885cb653795cda90480705e192c14144d7`. Logging sibling: `f913c480175d88b31adbefa5825d1283ac2da2f2`. The logging sibling diff adds only the requested environment-controlled append hook.','',
'One adapted tree was built with base stage3/apply.sh from TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8` and the base lockfile API packages. Both ordinary compilers share the same pinned cohere/checker dependency. Provenance and source hashes are saved beside the logs.','',
'Read and reused the bounded first-error selector and validation design from typescript/any-members-census at `f45740c18bccdbe5b87b7d3d5d576fb79002dce9`. The archived scripts in tooling retain their historical any-members paths and must not be run as container-fit launchers. build.py and run_piece.py are this task\'s scripts. The helper adds unit markers for attribution without changing walk logic; no logging measurement was reached.','',
'The base latent binary built, but the logging head overlay generator stopped while generating typed snapshots. This is a tooling build failure, not a measured compiler panic. containerFitPair is the new named `[2]*checker.Type` map key. The first-error overlay also generates these snapshot helpers, so LATENT_FULL=0 does not avoid the guard. The guard requires an explicit copier; this report does not bypass it or claim unsupported snapshots preserve behavior. The ordinary census does not use this overlay.','',
'Each launched measurement has a 1,200 second child limit, followed by at most 10 seconds before forced termination. One invocation of run_piece.py handles one piece, then commits and ordinarily pushes its saved output before returning. No long-lived multi-piece supervisor is used.','',
'## First stops','',
'| Requested cohort | Result |','| --- | --- |','| 20 utilitiesPublic.ts find sites | Unknown: no paired latent observations. None can be classified gone, moved, or unchanged. |','| Reported 50 binder.ts:427:31 latent roots | Unknown: no paired latent observations. None can be classified gone, moved, or unchanged. |','',
'The requested byte-identical head A versus logging-sibling utilitiesPublic piece check was not run because the logging overlay could not build.','',
'## Regressions and cap','',
'Latent finding-free unit and site regressions, changed first stops in already blocked units, and new latent panics are unknown. Main stage3/census/pair requires two complete compatible streams; neither stream exists here, so no synthetic pairing was attempted.','',
'Rows with `a fit walk past ... distinct pairs` cannot be counted in the unavailable latent census. Whether N=4096 is too low there is unknown. Any observed ordinary cap row is listed below.','',
'## Pair counts','',
'No walk log exists. Largest judged and past, totals, p50/p90/p99/max, largest kept per unit, and the five largest walks are all unknown.','',
'## Time','', 'Box: cloud host 6feab438e03f; CPU quota 400000/100000 us (4 cores); memory limit 17179869184 bytes (16 GiB). Go 1.27.1 linux/amd64; Node v24.19.0.','',
'All available times are one sample on this box. LATENT_TIMES sums over common units and the 20 biggest changes in each direction are unavailable because no paired latent runs occurred. Ordinary wall times include checker and lowering work and must not be substituted for latent unit times.','',
'| Ordinary 79-root run | Complete roots saved | Process wall seconds | Sum of per-root seconds |','| --- | ---: | ---: | ---: |']
for label,rs in [('base',b),('head',h)]:
 m=metrics.get(label,{})
 wall=f'{m["wall_seconds"]:.6f}' if 'wall_seconds' in m else 'unknown'
 lines.append(f'| {label} | {len(rs)} | {wall} | {sum(r["seconds"] for r in rs.values()):.6f} |')
common=sorted(set(b)&set(h));binder=[k for k in common if norm(b[k].get('where',''))=='src/compiler/binder.ts:427:31']
lines+=['','## Ordinary first-stop distribution','','| Kind and first stop | Base roots | Head roots |','| --- | ---: | ---: |']
counts={label:collections.Counter((r['kind'],norm(r.get('where','')),r.get('reason','')) for r in rs.values()) for label,rs in [('base',b),('head',h)]}
for key in sorted(set(counts['base'])|set(counts['head'])):
 label='accepted' if key[0]=='accepted' else f'{key[0]} {key[1]}: {key[2]}'
 lines.append(f'| {label} | {counts["base"][key]} | {counts["head"][key]} |')
lines+=['','## Ordinary binder root first stops','',f'{len(binder)} common roots have the requested base binder stop. These are ordinary per-file results, not latent attempting units.','', *(['| Root | Base stop | Head stop | Classification |','| --- | --- | --- | --- |'] if binder else [])]
stop=lambda r:f'{r["kind"]} {norm(r.get("where",""))}: {r.get("reason","")}'.strip()
for k in binder:
 br,hr=b[k],h[k];same=(br['kind'],norm(br.get('where','')),br.get('reason'))==(hr['kind'],norm(hr.get('where','')),hr.get('reason'))
 lines.append(f'| {k} | {stop(br)} | {stop(hr)} | {"unchanged" if same else "gone" if hr["kind"]=="accepted" else "moved"} |')
regressions=[dict(root=k,base=b[k],head=h[k]) for k in common if b[k]['kind']=='accepted' and h[k]['kind']!='accepted']
changed=[dict(root=k,base=b[k],head=h[k]) for k in common if b[k]['kind']!='accepted' and (b[k]['kind'],norm(b[k].get('where','')),b[k].get('reason'))!=(h[k]['kind'],norm(h[k].get('where','')),h[k].get('reason'))]
cap=[dict(label=label,root=k,row=r) for label,rs in [('base',b),('head',h)] for k,r in rs.items() if 'a fit walk past ' in r.get('reason','') and ' distinct pairs' in r.get('reason','')]
lines+=['','## Ordinary regression ledger','',f'Compared {len(common)} common roots. Finding-free to blocked ordinary roots: {len(regressions)}. Changed stops in already blocked roots: {len(changed)}. Cap rows: {len(cap)}. These counts cover only saved ordinary rows. Process failures and incomplete coverage remain in metrics and logs.']
for label,m in metrics.items():
 lines+=['',f'Ordinary {label} process exit: {m["returncode"]}; timed out: {m["timed_out"]}.']
 if m['returncode']==0:lines+=['No fatal ordinary census panic was observed in this completed run.']
for r in regressions:lines+=['',f'- {r["root"]}: {stop(r["head"])}']
for r in cap:lines+=['',f'**N IS TOO LOW HERE:** {r["label"]} {r["root"]}: {stop(r["row"])}']
lines+=['','Complete ordinary root rows and changes are in ordinary-summary.json. The build failure and process logs are retained in evidence/.','']
(O/'REPORT.md').write_text('\n'.join(lines))
(O/'ordinary-summary.json').write_text(json.dumps(dict(base=b,head=h,regressions=regressions,changed_blocked_roots=changed,cap=cap,metrics=metrics),indent=2)+'\n')
(O/'outcome.json').write_text(json.dumps(dict(schema=1,latent_status='stopped at overlay build',requested_utilities_find_sites=20,requested_binder_roots=50,latent_first_stops=None,latent_regressions=None,latent_cap_rows=None,pair_counts=None,latent_time_sums=None,head_log_byte_identity=None,ordinary_summary='ordinary-summary.json'),indent=2)+'\n')
