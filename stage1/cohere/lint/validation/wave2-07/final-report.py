import json,pathlib,subprocess,datetime,collections,re,gzip,shutil
root=pathlib.Path.cwd();gate=pathlib.Path('/tmp/wave2-07-gate');summary=json.loads((gate/'summary.json').read_text());meta=summary['metadata'];merges=json.load(open('/tmp/wave2-07-merges.json'));new=json.load(open('/tmp/wave2-07-new-rules.json'));out=pathlib.Path('/tmp/wave2-07-report');out.mkdir(exist_ok=True)
lines=['# Lint batch wave2-07 pre-gate','',f"Tested tree: `{meta['head']}`. Base: `f179cfb4bbacdc54911fcf3b5ec365b7c24e6914`.",'','## Merge scope','', 'The first two merges are the authorized shared exceptions: nonprogressing-fix-time `6ca65dbec`, then shards-program-line `eaddd5f73`. Rule/helper merges follow in the requested order, with two parents each. No rebase or squash. The explicitly authorized capturedTypedCases count changed from 373 to 415 for @typescript-eslint/prefer-reduce-type-parameter, after checking upstream RunTypedFiles tests. No strict-alone cases were added. The owned Next document oracle adapter was formatted.','', '| Candidate | Pinned commit | Result |','| --- | --- | --- |']
for r in merges:lines.append(f"| {r['branch']} | {r['sha']} | {r['status']} |")
lines+=['','### Exclusions','']
for r in merges:
 if r['status']!='merged':lines.append(f"- {r['branch']} `{r['sha']}`: "+'; '.join(r.get('reasons',[])))
lines+=['','## Gate inputs and commands','',f"nproc {meta['nproc']}; cpu.max `{meta['quota']}`. TypeScript `{meta['typescript_sha']}` at `{meta['inputs']['ADAMIC_TYPESCRIPT_SOURCE']}`; porcelain status including ignored files empty before and after. Fresh common profile directory `{meta['inputs']['ADAMIC_LINT_PROFILE_DIR']}`. WASI sysroot `{meta['inputs']['WASI_SYSROOT']}`. Lint and parser benchmark flags and ADAMIC_GATE_UNCACHED are 1. GOPROXY uses pipe fallback.",'','Each package ran `go test <package> -count=1 -v -json -timeout=3h`. All 95 changed Go files were checked by gofmt -l and go vet in their correct package/module context; no formatting residue or unused imports. See vet-coverage.json for the adapters and detached proof fixtures.','', '## Packages','', '| Package | Pass/fail/skip (all tests) | Pass/fail/skip (top level) | Test elapsed | Command wall |','| --- | --- | --- | --- | --- |']
for r in summary['results']:
 match=next((m for m in meta['packages'] if m['log']==r['log']),None);c=r['counts'];tc=r['top_counts'];end=[]
 for l in open(r['log']):
  try:e=json.loads(l)
  except:continue
  if e['Action'] in ['pass','fail','skip'] and 'Test' not in e:end.append(e)
 lines.append('| '+str(r['package'])+' | '+ '/'.join(str(c.get(a,0)) for a in ['pass','fail','skip'])+' | '+ '/'.join(str(tc.get(a,0)) for a in ['pass','fail','skip'])+' | '+str(end[-1].get('Elapsed','') if end else '')+'s | '+(f"{match['wall']:.3f}s" if match else 'in progress')+' |')
 for s in r['skips']:lines.append(f"\nSkip: `{s}`. See package log for its explicit reason.")
lint=next((r for r in summary['results'] if r['package'].endswith('/lint')),None)
if lint:
 lines+=['','## Lint serial phase and top-level test times','',f"Serial phase (package start to last serial test finish): {lint.get('last_serial_finish_seconds')}s, ending with {lint.get('last_serial_test')}. First parallel CONT: {lint['serial_seconds']}s. Seat baseline whole wall: 3,421.6s at f44d1b88. This batch includes 67 additional rules. This is a different box and a failing run; wall differences do not establish a green-gate speedup.",'','| Top-level test | Own time | Result |','| --- | --- | --- |']
 lines += [f"| {t['test']} | {t['seconds']:.2f}s | {t['result']} |" for t in lint['top_tests']]
if lint and lint.get('unfinished_top_tests'):
 lines += ['', 'Top-level tests without a terminal event: '+', '.join('`'+n+'`' for n in lint['unfinished_top_tests'])+'. Their final own time could not be measured from the requested events.']
lines += ['', 'Go reports TestMutants own time as 0.19s; its parallel children span 4,665.819s from parent CONT to terminal PASS. These are different measurements, and the complete JSON events are retained.']
lines+=['','## Registered rules and mutant catches','', '166 registered rules, 99 at base and 67 added. New rules and owner commits are in new-rules.json. Full per-backend catch messages are in mutant-catches.json. A missing catch is reported as missing, never inferred from another rule.','']
catch={r['package']:r['mutant_catches'] for r in summary['results']};(out/'mutant-catches.json').write_text(json.dumps(catch,indent=2))
fail={}
for r in summary['results']:
 for l in open(r['log']):
  try:e=json.loads(l)
  except:continue
  if e.get('Test') in r['failures'] and 'Output' in e:fail.setdefault(e['Test'],[]).append(e['Output'])
(out/'failure-output.json').write_text(json.dumps({k:''.join(v) for k,v in fail.items()},indent=2))
lines += ['', '## Failure evidence', '', 'Complete failure messages, including test source locations, are preserved in failure-output.json and the compressed package logs. Branch attribution, test file:line, messages and minimized reproducers are in [FAILURES.md](FAILURES.md).']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n');print(out)
