import json,os,pathlib,re,subprocess,time
out=pathlib.Path(__file__).resolve().parent
root=out.parents[2];cli='/tmp/loom-r-implied-target-adamic'
names=['invariance_readonly','optional_widening_class','optional_widening_declared','optional_widening_fresh','node_options_widening']
recorded={line.split(' | ')[0].removeprefix('| '):line for line in (root/'internal/oracle/counts.md').read_text().splitlines() if line.startswith('| ')}
results=[]
for name in names:
 fixture='internal/oracle/testdata/'+name+'.a';binary='/tmp/loom-r-implied-target-count-'+name
 command=[cli,'build',fixture,'-o',binary,'--count'];start=time.monotonic()
 build=subprocess.run(command,cwd=root,capture_output=True,timeout=90)
 (out/'logs'/('count-build-'+name+'.log')).write_bytes(build.stdout+build.stderr)
 if build.returncode:raise SystemExit('counted build failed: '+name)
 command=['/bin/sh','-c','ulimit -s 8192 && exec "$0"',binary]
 actual=subprocess.run(command,cwd=root,capture_output=True,timeout=30)
 (out/'logs'/('count-run-'+name+'.log')).write_bytes(actual.stdout+actual.stderr)
 match=re.search(rb'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$',actual.stderr)
 assert match,(name,actual.stderr)
 measured=[int(v) for v in match.groups()];expected=[int(v.strip().rstrip('|').strip()) for v in recorded[fixture].split(' | ')[1:]]
 row=dict(fixture=fixture,recorded=expected,measured=measured,exit=actual.returncode,seconds=time.monotonic()-start,match=measured==expected)
 results.append(row);(out/'direct-counts.json').write_text(json.dumps(results,indent=2)+'\n');print(name,row['match'],round(row['seconds'],3),flush=True)
 assert row['match'] and actual.returncode==0
