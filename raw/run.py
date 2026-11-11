import json, pathlib, subprocess, sys, time
out=pathlib.Path(sys.argv[1]); command=sys.argv[2:]; started=time.monotonic()
with out.with_suffix('.log').open('w') as stream:
    code=subprocess.run(command, stdout=stream, stderr=subprocess.STDOUT).returncode
result=dict(command=command,child_exit=code,seconds=time.monotonic()-started)
out.with_suffix('.status.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
raise SystemExit(code if code>=0 else 128-code)
