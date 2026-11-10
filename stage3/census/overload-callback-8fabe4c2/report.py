import html
import gzip,hashlib,json,pathlib,shlex
out=pathlib.Path(__file__).parent
j=lambda name:json.loads((out/name).read_text())
def esc(s):return html.escape(str(s),quote=False).replace('|','\\|').replace('\n','<br>')
def where(s):return 'src/'+s.split('/src/',1)[1] if '/src/' in s else s
def finding(f):return 'None' if not f else '"'+esc(f['kind']+': '+f['reason'])+'" at '+where(f['where'])
def stream(name):
 with gzip.open(out/(name+'.jsonl.gz'),'rt') as f:return [json.loads(x) for x in f]
prov=j('provenance.json')
lines=['# Paired latent census: overload callback use','','Task 8n7eet3, Optimizer request, November 10, 2026. Report only.','','Compiler parent: `'+prov['parent']['sha']+'`. Compiler head: `'+prov['head']['sha']+'`. The optimizer branch was checked before work and still pointed at this head. It had not moved. Each binary uses its own SHA and guarded latent overlay, extended by its own `stage3/meter/entry_overlay.py`. No compiler or adaptation source was changed.','','One adapted tree was built once by parent `1d84074f` using `stage3/apply.sh`, from TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both sides use those same adapted bytes. `source-hashes.json` freezes all 747 source files; `patch-set.md`, `provenance.json`, `overlay-hashes.json`, and the raw child commands record reproduction inputs. CoHere and its checker dependency match on both sides.','','Both entry drivers require `LATENT_ASSERT_NO_OUTPUT=1`, check that ordinary Load and Lower cannot expose output, and import no backend. Measurements concern independent lowering units, not native behavior or whole-program compilation. Deferred generic declarations are retained as deferred, not counted as successful standalone lowerings.','','## Run resources','','Wall time and process peak RSS come from each census child\'s own completed resource runner. Both sides overlap on this host and use `GOMEMLIMIT=5GiB`; these are resource observations, not a speed benchmark.','','| Run | Child exit | Wall seconds | User CPU seconds | System CPU seconds | Peak RSS KiB | Peak RSS GiB |','| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for name in ['debug-parent','debug-head','entry-parent','entry-head']:
 p=out/(name+'.status.json')
 if p.exists():
  s=j(p.name);lines.append(f"| {name} | {s['child_exit']} | {s['wall_seconds']:.3f} | {s['user_seconds']:.3f} | {s['system_seconds']:.3f} | {s['peak_rss_kib']} | {s['peak_rss_kib']/1048576:.3f} |")
 else:lines.append('| '+name+' | running | pending | pending | pending | pending | pending |')
lines+=['', (out/'debug-section.md').read_text(), '## tsc entry, first-error mode','']
pair=out/'entry-paired.json'
if pair.exists():
 d=j(pair.name)
 lines+=['Both first-error streams load the same checker-clean tsc entry and resolved 81-file graph. LATENT_FULL, LATENT_SPECULATIVE and LATENT_ONLY_FILE are unset. The pairing tool verified equal headers, unit identities and eligibility, and normalized locations to the outermost contiguous property/element access chain.','','A cleared finding below leaves no remaining finding at that identity. A changed reason is a removed reason with another reason still at the identity. New findings include exposed reasons at already-blocked sites; new sites have no parent finding. Changed first stops in already-blocked units are reported separately from regressions.','','| Measure | Raw | Distinct |','| --- | ---: | ---: |']
 for title,key in [('Cleared, no remaining finding','lowered'),('Changed reason, still blocked','moved'),('New findings','new_findings'),('New sites','new_sites'),('All unit first-stop changes','unit_first_stop_delta'),('Regressions','regressions')]:
  lines.append(f"| {title} | {len(d['raw_findings'][key])} | {len(d['distinct_reads'][key])} |")
 for title,fn in [('All removed findings/reasons',lambda v:len(v['lowered'])+len(v['moved'])),('Already-blocked units with a changed first stop',lambda v:sum(bool(r['base']) and bool(r['head']) for r in v['unit_first_stop_delta'])),('Previously blocked units now finding-free',lambda v:sum(bool(r['base']) and not r['head'] for r in v['unit_first_stop_delta']))]:
  lines.append(f"| {title} | {fn(d['raw_findings'])} | {fn(d['distinct_reads'])} |")
 lines+=['','| Finding kind | Raw parent | Raw head | Distinct parent | Distinct head |','| --- | ---: | ---: | ---: | ---: |']
 kinds=sorted(set(d['raw_findings']['totals']['before'])|set(d['raw_findings']['totals']['after']))
 for k in kinds:
  lines.append('| '+k+' | '+' | '.join(str(d[n]['totals'][side].get(k,0)) for n,side in [('raw_findings','before'),('raw_findings','after'),('distinct_reads','before'),('distinct_reads','after')])+' |')
 units={tuple(r['identity']):r for n in ['distinct_reads','raw_findings'] for r in d[n]['unit_first_stop_delta']}
 lines+=['','Every unit whose first stop changes, with whole reasons:','','| Unit | Declaration | Parent first stop | Head first stop | Classification |','| --- | --- | --- | --- | --- |']
 for key,r in sorted(units.items()):
  u=r['unit'];name=u.get('name') or u['kind'];classification='already blocked, changed stop' if r['base'] and r['head'] else ('cleared unit' if r['base'] else 'regression: newly stopped unit')
  lines.append('| '+esc(name)+' | '+where(u['where'])+' | '+finding(r['base'])+' | '+finding(r['head'])+' | '+classification+' |')
 if not units:lines.append('| None | | | | |')
 outside=[r for r in units.values() if not where(r['unit']['where']).startswith('src/compiler/debug.ts:') and any(f and 'overload ' in f['reason'] and 'parameter ' in f['reason'] and 'cannot be served by implementation parameter' in f['reason'] for f in [r['base'],r['head']])]
 lines+=['','Outside-debug units with a moving overload-parameter reason (all quoted whole, including newly exposed stops):','']
 for r in sorted(outside,key=lambda r:r['unit']['where']):
  u=r['unit'];lines.append('- `'+where(u['where'])+'` (`'+(u.get('name') or u['kind'])+'`): '+finding(r['base'])+' -> '+finding(r['head'])+'.')
 if not outside:lines.append('None.')
 lines+=['','### Regressions','','A regression is a previously finding-free attempted unit that now stops, a newly finding-bearing scan site, or a new panic. Exposed first stops in already-blocked units are excluded.','']
 regressions={json.dumps(r,sort_keys=True):r for r in d['distinct_reads']['regressions']}
 lines+=['Regression conclusions use the normalized distinct-read identities. The raw column retains unnormalized candidate counts from the pairing tool.','']
 for r in regressions.values():
  f=r.get('finding') or r.get('head');lines.append('- '+r['type']+': '+finding(f)+(' in unit '+where(r['unit']['where']) if r.get('unit') else '')+'.')
 if not regressions:lines.append('**None.**')
 lines+=['','Panics: '+str(len(d['raw_findings']['panics']['before']))+' parent, '+str(len(d['raw_findings']['panics']['after']))+' head. The exact cleared, changed, new and regression records are retained in [entry-paired.json](entry-paired.json) and [entry-paired.md](entry-paired.md).','']
else:lines+=['The two entry runs are in progress. No completed entry delta or regression conclusion is available yet.','']
lines+=['## Evidence, checks and checkpoints','','The 18 existing access-chain pairing tests passed. Both production-tree comparisons are clean in the compiler, adapter, meter and pairing directories. Every completed census must have child exit 0, a measurement header, zero checker diagnostics and a final relation-report trailer; [verification.json](verification.json) records the final recomputation of adapted source and archive hashes.','','The initial environment transport reconnect ended the original outer setup/build/apply wrappers. The builds and apply continued with unchanged arguments and their own exit 0 was recovered from Linux kernel wait status by adopt.py. Setup was retried and completed with exit 0. A second transport reconnect during the entry runs left both detached supervisors and their children alive with unchanged commands; entry-reconnect.json records their identities. Only the preparation statuses use recovered process statistics; all four census runs use direct-child resource measurement under detached supervisors. A 30-minute period without log or watched-file progress stops a child process group. No stopped run is treated as complete.','','Automatic approval review initially rejected the apply checkpoint push as an unverified separate repository. Inspection established that it targets another branch of the same authorized origin, https://github.com/system-inc/adamic.git. The explicit same-repository retry was approved and succeeded.','','Checkpoint snapshots are on `typescript/overload-callback-census-checkpoints` in that same repository. The final requested branch `typescript/overload-callback-census` has one report-only commit directly on `1d84074f`, touching only this directory. No main push or force-push is used. Checkpoint push logs are retained beside the report.','','Raw streams are compressed with deterministic gzip even when below 20 MB. The following hashes cover the original decompressed stream and the retained archive; each run\'s `*-streams.json` also records timing-stream hashes.','','| Raw stream archive | Original bytes | Original SHA-256 | Archive bytes | Archive SHA-256 |','| --- | ---: | --- | ---: | --- |']
for name in ['debug-parent','debug-head','entry-parent','entry-head']:
 p=out/(name+'-streams.json')
 if p.exists():
  for r in j(p.name):
   if r['file']!=name+'.jsonl.gz':continue
   lines.append(f"| [{r['file']}]({r['file']}) | {r['bytes']} | `{r['sha256']}` | {r['compressed_bytes']} | `{r['compressed_sha256']}` |")
lines+=['','[commands.md](commands.md) records reproduction commands. Exact census commands, flags, exits and resources are in the four `*.status.json` files. Setup used `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and `ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-typescript` at the required upstream pin. The parent applied once to `/workspace/overload-adapted`. Each compiler then generated its own latent overlay and entry extension and built with `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`.','']
text='\n'.join(lines)
assert '\u2014' not in text and '\u2013' not in text
(out/'REPORT.md').write_text(text)
print('report bytes',len(text.encode()),'entry paired',pair.exists())
