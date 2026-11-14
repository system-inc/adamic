import json,pathlib,subprocess
s=pathlib.Path('/workspace/scratch/path-mint-618'); pins=json.loads((s/'work/stage3/adapt/99-correctness-timer-closures/owners.json').read_text())
for f in sorted({p['file'] for p in pins}):
 target=s/'contract-before'/f;target.parent.mkdir(parents=True,exist_ok=True)
 target.write_bytes(subprocess.check_output(['git','show','refs/stage3/99-correctness-timer-closures-before:'+f],cwd=s/'base'))
