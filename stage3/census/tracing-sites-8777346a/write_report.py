import json,hashlib,shutil,collections
from pathlib import Path
root=Path('/workspace/tracing-sites-8777346a'); scratch=Path('/tmp/tracing-8777346a'); out=root/'stage3/census/tracing-sites-8777346a';out.mkdir(parents=True,exist_ok=True)
rows=[json.loads(l) for l in (scratch/'raw.jsonl').read_text().splitlines()]
result=json.loads((scratch/'census-result.json').read_text())
assert result['exit_code']==0
assert rows[0]['diagnostics']==[] and rows[0]['diagnostic_sites']==[] and rows[0]['checker_rejected']==False and rows[0]['latent_mode']=='full'
assert rows[-1]['status']=='relation_report'
source=scratch/'adapted/src/compiler/tracing.ts'; lines=source.read_text().splitlines(); file=str(source)
findings=[f for r in rows for f in r.get('findings',[])]
by={}
expressions={66:'require("fs")',98:'tracingEnabled',166:'eventStack.length--',173:'eventStack.length = 0',185:'...args',196:'args',340:'fs as { writeSync(fd: number, data: string | undefined): number }',359:'fs as { writeFileSync(path: string, data: string | undefined): void }',380:'declare const require: (specifier: "fs") => any;'}
for f in findings:
 if f['kind'] not in ('NotYet','Refused') or f['where'].rsplit(':',2)[0]!=file:continue
 key=tuple(f[k] for k in ('kind','where','reason','text'))
 if key not in by:
  line,col=map(int,f['where'].rsplit(':',2)[1:]); expr=f['reason'][8:] if f['reason'].startswith('reading ') else expressions[line]
  by[key]=dict(file='src/compiler/tracing.ts',line=line,column=col,location=f'{line}:{col}',kind=f['kind'],reason=f['reason'],source_expression=expr,source_line=lines[line-1],phases=[],occurrences=[])
 by[key]['phases']+=[] if f['phase'] in by[key]['phases'] else [f['phase']]
 by[key]['occurrences'].append(dict(phase=f['phase'],unit=f['unit'],text=f['text']))
sites=sorted(by.values(),key=lambda s:(s['line'],s['column'],s['reason']))
checks=[]
for line,expr in [(311,'recursionIdentityMap.get(recursionIdentity)'),(313,'recursionIdentityMap.size'),(314,'recursionIdentityMap.set(recursionIdentity, recursionToken)'),(340,expressions[340]),(359,expressions[359]),(380,expressions[380]),(381,'declare const process: { readonly pid: number };')]:
 needle='recursionIdentityMap' if line in (311,313,314) else ('fs as' if line in (340,359) else 'declare')
 col=lines[line-1].index(needle)+1
 matches=[s for s in sites if s['line']==line]
 checks.append(dict(line=line,column=col,source_expression=expr,status='reported' if matches else 'absent from the stream',findings=[dict(kind=s['kind'],reason=s['reason'],phases=s['phases'],location=s['location']) for s in matches]))
assert len(sites)==16 and len(checks)==7
assert all(c['status']=='absent from the stream' for c in checks[:3])
provenance=dict(base='8777346abe4d888c0182e301b1ffec76bd45ae04',source_sha256=hashlib.sha256(source.read_bytes()).hexdigest(),source_bytes=source.stat().st_size,source_path=str(source),environment=dict(LATENT_FULL='1',LATENT_ONLY_FILE=str(source),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(scratch/'unit-times.jsonl')),other_LATENT_variables=[],child_result=result,checker_diagnostics=0,header=rows[0],deduplication=['kind','where','reason','text'],boundary_markers_excluded=21,dependency_findings_excluded=7)
(out/'sites.json').write_text(json.dumps(dict(provenance=provenance,distinct_site_count=len(sites),sites=sites,requested_checks=checks),indent=2)+'\n')
for name in ['raw.jsonl','unit-times.jsonl','run.py','census.log','census-result.json','apply-retry.log','apply-retry-result.json','overlay-final.log','overlay-final-result.json','build.log','build-result.json','audit.log','audit-result.json','full-audit.log','full-audit-result.json','gofmt-result.json']:
 shutil.copyfile(scratch/name,out/name)
shutil.copyfile(source,out/'tracing.ts.txt')
# Preserve generated Go evidence with .go.txt names, plus a byte-hash manifest.
evidence=out/'overlay';evidence.mkdir(exist_ok=True)
manifest={}
for p in sorted((scratch/'overlay-final').glob('*.go')):
 shutil.copyfile(p,evidence/(p.name+'.txt'));manifest[p.name]=hashlib.sha256(p.read_bytes()).hexdigest()
(out/'overlay-sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
shutil.copyfile(scratch/'overlay-final/overlay.json',evidence/'overlay.json')
(out/'write_report.py').write_bytes(Path(__file__).read_bytes())
md=['# tracing.ts latent census at 8777346a','',f'Base: `{provenance["base"]}`. Branch: `typescript/tracing-sites-8777346a`. This is the single-file full-mode census requested for Optimizer task #ff5eddz.','',f'The census child exited **0** in {result["seconds"]:.3f} seconds, with **zero checker diagnostics**. The stream header says `latent_mode: full`, `checker_rejected: false`, and empty `diagnostics` and `diagnostic_sites`. The final `relation_report` is present. The overlay\'s fixed measurement label still says “measured on a checker-rejected program”; the actual header establishes that this run is checker-clean.','', '**16 distinct tracing.ts sites: 13 NotYet and 3 Refused**, comprising 18 phase-specific observations. The two method views occur in both phases. Deduplication follows the README: `(kind, where, reason, text)`; each row retains all observed phases. Attribution uses the finding\'s `where`, not the attempting unit: seven findings in checker.ts are not tracing.ts sites. The 21 Boundary markers are excluded from this ledger; raw.jsonl preserves the complete stream. All 23 catalogued units were attempted.','', '## Requested checks','']
for c in checks:
 if not c['findings']: md.append(f'- `{c["line"]}:{c["column"]}` — `{c["source_expression"]}`: **absent from the stream** (neither a NotYet nor a Refused finding).')
 else:
  s=c['findings'][0];md.append(f'- `{s["location"]}` — `{c["source_expression"]}`: **{s["kind"]}**, {", ".join(s["phases"])}; exact reason: {json.dumps(s["reason"])}.')
md+=['','The three map reads remain in the source; their census findings are absent. The map declaration is now at line 226 and is also absent as a site. The old line numbers are not used to infer current findings. Absence is an observation from this full-mode stream, not proof that every child expression was examined or that whole-program lowering succeeds. Full mode can stop at failed constructs, and boundary markers document those limits.','','## Complete site ledger','','Columns are one-based UTF-16 positions as emitted by the census. Source expressions are copied from the adapted file; the complete source line is also in sites.json. Reasons below are the exact `reason` field, without paraphrase.','','| Line:column | Phase | Kind | Exact reason | Source expression |','| --- | --- | --- | --- | --- |']
for s in sites:
 esc=lambda t:t.replace('|','\\|').replace('\n',' ')
 md.append(f'| {s["location"]} | {", ".join(s["phases"])} | {s["kind"]} | {esc(s["reason"])} | `{esc(s["source_expression"])}` |')
md+=['','## Reproduction and evidence','',f'Adapted source: `{source}`; SHA-256 `{provenance["source_sha256"]}` ({provenance["source_bytes"]} bytes). `tracing.ts.txt` preserves those exact bytes. TypeScript source pin: `050880ce59e30b356b686bd3144efe24f875ebc8` (v6.0.3). The worktree started at the exact base above; no compiler sources were modified. Recursive submodules and the stage3/api lockfile packages were initialized.','', 'Commands from the pinned worktree, with `/workspace/adamic-tools/env.sh` sourced and `ADAMIC_TYPESCRIPT_SOURCE=/tmp/tracing-8777346a/typescript-source`:','','```sh','bash stage3/apply.sh /tmp/tracing-8777346a/adapted','python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/tracing-8777346a/overlay-final','gofmt -w /tmp/tracing-8777346a/overlay-final/*.go','go build -buildvcs=false -overlay=/tmp/tracing-8777346a/overlay-final/overlay.json -o /tmp/tracing-8777346a/latent ./stage3/census/latent/tool','python3 stage3/census/latent/audit.py /tmp/tracing-8777346a/latent','python3 /tmp/tracing-8777346a/run.py census /tmp/tracing-8777346a/latent /tmp/tracing-8777346a/adapted/src/compiler /tmp/tracing-8777346a/raw.jsonl','```','', '`run.py` follows `stage3/adapt/99-correctness-axolotl-s/clear-census.py`: it removes every inherited LATENT_ variable and sets only the following for the census child:','','```json',json.dumps(provenance['environment'],indent=2),'```','', 'Every build/run wrapper writes its child\'s own exit code and exits with that code (signals map to 128 + signal). Each child has a 1740-second limit, followed by at most five seconds of termination grace. The successful apply took 638.587 seconds, overlay generation 7.759 seconds, build 241.012 seconds, basic audit 0.521 seconds, and census 42.688 seconds. No process was allowed to run for 30 minutes.','', 'The README basic audit passed. An additional `full_audit.py` run exited 1: its exact expected reason set is `{a string as a condition, a number as a condition, var}`, but the actual set was `{var}`. This extra audit is **not claimed to pass**; its log and result are included. It does not alter the requested census result or its zero-diagnostic gate.','', 'The initial apply failed because ADAMIC_TYPESCRIPT_SOURCE was unset; initial overlay attempts failed because the temporary directory and nested checker submodule were missing. Those setup inputs were supplied, and the successful build results above supersede those attempts.','', 'Evidence includes raw.jsonl, unit-times.jsonl, the bounded wrapper, build/audit/census logs and child results, the exact adapted source, and the generated scratch overlay. Every Go evidence file is named `.go.txt`. `sites.json` retains every occurrence, phase, exact reason, and source expression. The generated overlay retains its original absolute-path mapping for provenance; its paths must be regenerated when reproducing elsewhere.','']
(out/'REPORT.md').write_text('\n'.join(md))
print(json.dumps(dict(sites=len(sites),checks=checks,source_sha256=provenance['source_sha256']),indent=2))
