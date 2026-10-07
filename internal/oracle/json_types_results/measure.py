import os,subprocess,pathlib,json,time,sys
root=pathlib.Path('/workspace/adamic-json-types-scratch');runner=root/'oracle/node.mjs';after=runner.read_bytes();before=pathlib.Path('/tmp/json-types-original-node.mjs').read_bytes()
def output(args):return subprocess.check_output(args,cwd=root,text=True).strip()
metadata={'commit':output(['git','rev-parse','HEAD']),'unitCommit':output(['git','-C','/workspace/adamic','rev-parse','HEAD']),'nproc':output(['nproc']),'cpu.max':pathlib.Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'go':output(['go','version']),'clang':output(['clang','--version']).splitlines()[0],'node':output(['node','--version']),'mode':'ADAMIC_GATE_UNCACHED=1; Go package/build cache and native runtime archive cache warm; no observation cache; 16 Python CPU burners','beforeHelper':'go run ./oracle/json_types.go <fixture>','afterHelper':'go build -o <process-temp>/json_types ./oracle/json_types.go; execute helper for each fixture'}
rows=[];burners=[]
try:
 for _ in range(16):burners.append(subprocess.Popen([sys.executable,'-c','while True: pass'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL))
 for round in range(1,4):
  for variant in (['before','after'] if round%2 else ['after','before']):
   runner.write_bytes(before if variant=='before' else after)
   calls=pathlib.Path(f'/tmp/json-types-final-loaded-{round}-{variant}-go.jsonl');calls.write_text('')
   env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',JSON_GO_LOG=str(calls),PATH='/tmp/json-types-go-wrapper:'+os.environ['PATH'])
   cmd=['/tmp/json-types-'+variant+'.test','-test.run','^TestNativeAgreesWithNode/internal/oracle/testdata/json_(decode|encode)_','-test.count=1','-test.timeout=30m','-test.v']
   log=f'/tmp/json-types-final-loaded-{round}-{variant}.log';load=pathlib.Path('/proc/loadavg').read_text().strip();start=time.monotonic()
   with open(log,'w') as stream:p=subprocess.run(cmd,cwd=root/'internal/oracle',env=env,stdout=stream,stderr=subprocess.STDOUT)
   invocations=[json.loads(line) for line in calls.read_text().splitlines()];row=dict(metadata,round=round,variant=variant,exit=p.returncode,seconds=time.monotonic()-start,loadBefore=load,loadAfter=pathlib.Path('/proc/loadavg').read_text().strip(),buildRequests=sum(c[0] in ['run','build'] for c in invocations),command='ADAMIC_GATE_UNCACHED=1 JSON_GO_LOG='+str(calls)+' PATH=/tmp/json-types-go-wrapper:$PATH '+' '.join(cmd)+' > '+log+' 2>&1')
   rows.append(row);pathlib.Path('/tmp/json-types-final-loaded-timings.json').write_text(json.dumps(rows,indent=2));print(json.dumps(row),flush=True)
finally:
 runner.write_bytes(after)
 for p in burners:p.terminate()
 for p in burners:p.wait()
