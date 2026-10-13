import json, os, pathlib, shlex, subprocess, sys, time
root = pathlib.Path(__file__).resolve().parents[3]
out = pathlib.Path(__file__).resolve().parent
name, *command = sys.argv[1:]
started = time.monotonic()
with (out / (name + '.log')).open('w') as log:
    code = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT).returncode
record = {'name': name, 'command': shlex.join(command), 'exit': code, 'seconds': time.monotonic() - started}
with (out / 'commands.jsonl').open('a') as ledger:
    ledger.write(json.dumps(record) + '\n')
print(json.dumps(record), flush=True)
sys.exit(code)
