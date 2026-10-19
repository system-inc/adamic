import json,subprocess,os,time
from pathlib import Path
root=Path.cwd();out=Path(__file__).resolve().parent; results=[]
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
os.environ["GOMAXPROCS"]="4"
cases=[('fixture-owner-omitted','stage3/fixtures/self_compare_test.go','TestFixturesSelfCompareClosure','checkFixturesSelfCompareClosure','./stage3/fixtures','^TestFixtureDirectoriesHaveTopLevelTests$','has no top-level test'),('reader-allowlist-omitted','internal/ir/call_targets_guard_test.go','"internal/native/arguments_length.go:spreadArguments:Call.Function":                               {"runtime", "names the spread call signature"},','', './internal/ir','^TestCallTargetReadersNative$','unapproved call-target read')]
for name,path,before,after,pkg,pattern,marker in cases:
 source=(root/path).read_text();assert source.count(before)==1
 folder=out/name;folder.mkdir(exist_ok=True);planted=folder/'source.go.txt';planted.write_text(source.replace(before,after))
 overlay=folder/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(planted)}}))
 command=['timeout','180','go','test',pkg,'-overlay='+str(overlay),'-run',pattern,'-json','-count=1','-timeout=90s']
 start=time.monotonic()
 original = (root/path).read_bytes()
 try:
  # The ownership guard parses physical source, so a Go overlay alone cannot
  # mutate its observation. Restore the source even if the child fails.
  if name == 'fixture-owner-omitted': (root/path).write_text(source.replace(before,after))
  with (folder/'test.jsonl').open('w') as f:ran=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,timeout=190)
 finally:
  if name == 'fixture-owner-omitted': (root/path).write_bytes(original)
 text=(folder/'test.jsonl').read_text();caught=ran.returncode==1 and marker in text and 'test timed out' not in text and 'build-fail' not in text
 row=dict(mutant=name,caught=caught,exit=ran.returncode,seconds=time.monotonic()-start,command=command,marker=marker);results.append(row)
 (out/'grain-mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True);assert caught
