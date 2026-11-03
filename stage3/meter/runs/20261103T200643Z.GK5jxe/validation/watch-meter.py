import json, pathlib, re, subprocess, time
root=pathlib.Path('/workspace/scratch')
cg=pathlib.Path('/sys/fs/cgroup')
trace=root/'meter-exit-trace.txt'
start=(root/'meter-memory-samples.jsonl').stat().st_ctime
# The original sampler's first sample gives the wall-clock run origin via birth time.
start=__import__('os').stat(root/'meter.log').st_mtime if False else 1793736393.038591
with (root/'meter-observer.log').open('w') as log:
    observer=subprocess.Popen(['strace','-e','trace=exit_group','-p','10741','-o',str(trace)],
                              stdout=log,stderr=log)
    with (root/'meter-memory-samples-resumed.jsonl').open('w') as samples:
        while observer.poll() is None:
            samples.write(json.dumps({'observed_unix_seconds':time.time(),
                                      'memory_current_bytes':int((cg/'memory.current').read_text())})+'\n')
            samples.flush()
            time.sleep(1)
text=trace.read_text()
match=re.search(r'exit_group\((\d+)\)',text)
if not match:
    raise RuntimeError('Cannot determine meter timeout exit status: '+text+'; '+(root/'meter-observer.log').read_text())
code=int(match[1])
(root/'meter.exit').write_text(str(code)+'\n')
original=[json.loads(line)['memory_current_bytes'] for line in (root/'meter-memory-samples.jsonl').read_text().splitlines()]
resumed=[json.loads(line)['memory_current_bytes'] for line in (root/'meter-memory-samples-resumed.jsonl').read_text().splitlines()]
record={'exit_code':code,'exit_status_source':'strace exit_group of original timeout PID 10741; outer sampler interrupted by environment reconnect',
        'elapsed_seconds':time.time()-start,'sampled_peak_bytes':max(original+resumed),
        'sample_interval_seconds':1,'sampling_interruption':'2026-11-03T21:02:47Z until resumed observer start',
        'after':{name:(cg/name).read_text().strip() for name in ('memory.max','memory.peak','memory.events')}}
(root/'meter-resources.json').write_text(json.dumps(record,indent=2)+'\n')
