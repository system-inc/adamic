import gzip, hashlib, json, pathlib, shutil, subprocess
root=pathlib.Path('/workspace/adamic')
run=root/'stage3/meter/runs/20261103T200643Z.GK5jxe'
scratch=pathlib.Path('/workspace/scratch')
validation=run/'validation';validation.mkdir(exist_ok=True)
names=['setup.log','setup.exit','origin-main.sha','code.sha','meter.exit','meter-resources.json',
       'meter-exit-trace.txt','meter-observer.log','watch-meter.log','meter-tests.log','meter-tests.exit',
       'compiler-tests.log','compiler-tests.exit','report-tests.log','report-tests.exit','entry-tests.log',
       'entry-tests.exit','meter-go-tests.log','meter-go-tests.exit','measure-once-mutant-test.log',
       'measure-once-mutant.log','measure-once-mutant.exit','daily-mutants.log','daily-mutants.exit',
       'run-meter.py','watch-meter.py','summarize-meter.py']
for name in names:
    source=scratch/name
    if source.exists(): shutil.copyfile(source,validation/(name+'.txt' if name.endswith('.log') else name))
for name in ['meter.log','meter-memory-samples.jsonl','meter-memory-samples-resumed.jsonl']:
    with (scratch/name).open('rb') as source, gzip.open(validation/(name+'.gz'),'wb') as dest:
        shutil.copyfileobj(source,dest)
shutil.copyfile(root/'stage3/meter/twice-daily.sh',validation/'launcher.sh')
shutil.copyfile(scratch/'archive-meter.py',validation/'archive-meter.py')
summary=json.loads((run/'measurement-summary.json').read_text())
failures={key:summary[key] for key in ['latent_compiler_panics','latent_compiler_errors','latent_entry_panics','latent_entry_errors','ordinary_errors']}
(run/'failures-verbatim.json').write_text(json.dumps(failures,indent=2)+'\n')
checks={}
for name in ['census.jsonl.gz','latent.jsonl.gz','latent-times.jsonl.gz','tsc/census.jsonl.gz','tsc/latent.jsonl.gz']:
    a=gzip.decompress((run/'main'/name).read_bytes());b=gzip.decompress((run/'area'/name).read_bytes())
    assert a==b,name
    checks[name]={'identical':True,'uncompressed_sha256':hashlib.sha256(a).hexdigest()}
report=json.loads((run/'report.json').read_text())
assert report['trees']['main']==report['trees']['area']
assert len(summary['first_lowering_stops'])==4
assert sum(g['count'] for g in summary['first_lowering_stops'])==77
(validation/'alias-verification.json').write_text(json.dumps({'trees_identical':True,'streams':checks},indent=2)+'\n')
lines=['','## Measurement record','',
       'Measured origin/main: `'+summary['measured_commit']+'`. It contains `fa47f0d8`.',
       'Launcher code commit: `'+(scratch/'code.sha').read_text().strip()+'`.',
       'Daily command used the README environment and timeout, with STAGE3_METER_MAIN_REF set to the recorded origin/main SHA.',
       'Exit: 0. Wall: 4143 seconds (69m 3s). Cgroup limit: 17179869184 bytes (16 GiB).',
       'Cgroup memory.peak: 15942144000 bytes (%.6f GiB). No OOM or OOM kills.'%(15942144000/2**30),
       'One-second sampled peak: 15923269632 bytes. The environment reconnect interrupted the outer sampler; the original timeout and meter continued. The resumed observer attached strace to that timeout and recorded exit_group(0). Sampling has an explicitly recorded gap; the cgroup peak counter continued throughout.',
       '',
       '| Measure | Main and copied area |','| --- | ---: |',
       '| Whole-program checker diagnostics | 0 |','| tsc entry checker diagnostics | 0 |',
       '| Own-file diagnostics summed | 0 |','| Files past checker, own file | 79/79 |',
       '| Files past checker, whole program | 79/79 |','| Files lowering | 2/79 |',
       '| Compiler latent NotYet distinct sites | 560 |','| Compiler latent Refused distinct sites | 4294 |',
       '| Compiler latent panics / errors | 0 / 0 |',
       '| Entry latent NotYet distinct sites | 207 |','| Entry latent Refused distinct sites | 4240 |',
       '| Entry latent panics / errors | 2279 / 0 |',
       '| Entry resolved sources | 81 |','',
       'Both latent streams have 0 skipped dependencies. Counts deduplicate (kind, where, reason, text).',
       'The separate 79-root ordinary aggregate reports `lower: stage 0 compiles a program from one entry file, got 79`; its checker diagnostic count is still zero. Three JSON inputs are non-source input errors and excluded from the 79-file denominator.',
       '',
       '### First lowering stops for all 77 non-lowering source files','',
       '| Kind | Reason | Count | Owner from owners.json |','| --- | --- | ---: | --- |']
for group in summary['first_lowering_stops']:
    lines.append('| '+' | '.join([group['kind'],group['reason'].replace('|','&#124;'),str(group['count']),group['owner']])+' |')
lines += ['','Every affected file, exact first stop, and fix is listed in [measurement-summary.json](measurement-summary.json).',
          'Every panic and error reason/text is retained verbatim in [failures-verbatim.json](failures-verbatim.json), including all 2279 entry panic sites and the ordinary aggregate/non-source errors.',
          'The main and area report trees are identical; all five copied streams match byte for byte after decompression. See [alias verification](validation/alias-verification.json).',
          '',
          '### Validation','',
          'Compiler tests: 9 passed. Reporter tests: 25 ran, 2 optional binary probes skipped. Entry tests: 11 ran, 2 optional binary probes skipped.',
          'The second-measurement mutant is caught by `8 != 4 : main must be measured exactly once`.',
          'The full Python discovery ran 46 tests, with 4 skips and one pre-existing generator probe failure: its temporary probe cannot import github.com/system-inc/adamic/internal/ir. The broader daily mutant runner also stops at this unrelated generator failure after catching the launcher mutants.',
          'The exact requested `go test -count=1 ./stage3/meter/...` exits 1 with `go: warning: "./stage3/meter/..." matched no packages` and `no packages to test`. There are no Go tests reading twice-daily.sh or report.py.',
          'All command logs and exit codes are under [validation/](validation/). Setup exited 0. report.py is unchanged.',
          '']
with (run/'README.md').open('a') as out:out.write('\n'.join(lines))
print('archived',len(list(validation.iterdir())),'validation files')
