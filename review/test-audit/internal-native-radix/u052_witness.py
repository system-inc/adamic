import pathlib,json,subprocess,difflib,time
root=pathlib.Path.cwd();out=root/'review/test-audit/internal-native-radix';file=root/'internal/native/record_test.go';original=file.read_text()
weak=original.replace('stdout == want {','stdout == want || stdout != want {').replace('Flags(Options{Sanitize: true, Count: true})','Flags(Options{Sanitize: false, Count: true})')
a=weak.index('\texit, ok := err.(*exec.ExitError)',weak.index('func recordMemberStop('));b=weak.index('\n}',a)
weak=weak[:a]+'\treturn true'+weak[b:]
path=pathlib.Path('/tmp/u052-record-witness.go');path.write_text(weak)
overlay=pathlib.Path('/tmp/u052-witness-overlay.json');overlay.write_text(json.dumps({'Replace':{str(file):str(path)}}))
(out/'W01-W02.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),weak.splitlines(True),fromfile='a/internal/native/record_test.go',tofile='b/internal/native/record_test.go')))
cmd=['timeout','120','go','test','-json','-overlay',str(overlay),'-count=1','-timeout','90s','./internal/native/','-run','^(TestRecordReadMutants|TestRecordMutantsUnit.*)$']
start=time.monotonic()
with (out/'witness.log').open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
(out/'witness-run.json').write_text(json.dumps({'command':' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start},indent=2))
