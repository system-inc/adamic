import json, os, pathlib, subprocess, time
root = pathlib.Path('/workspace/scratch')
cg = pathlib.Path('/sys/fs/cgroup')
def snapshot():
    return {name: (cg/name).read_text().strip() for name in ('memory.max', 'memory.peak', 'memory.events')}
before = snapshot()
start = time.time()
with (root/'meter.log').open('w') as log, (root/'meter-memory-samples.jsonl').open('w') as samples:
    proc = subprocess.Popen(['timeout', '--kill-after=60s', '7200', 'bash', 'stage3/meter/twice-daily.sh'],
                            stdout=log, stderr=log, env=dict(os.environ, STAGE3_METER_COMPILER='per-ref', STAGE3_METER_LATENT='first-error',
                            STAGE3_METER_MAIN_REF=(root/'origin-main.sha').read_text().strip()))
    peak = 0
    while True:
        current = int((cg/'memory.current').read_text())
        peak = max(peak, current)
        samples.write(json.dumps({'elapsed_seconds':time.time()-start, 'memory_current_bytes':current})+'\n')
        samples.flush()
        if proc.poll() is not None:
            break
        time.sleep(1)
    code = proc.wait()
(root/'meter.exit').write_text(str(code)+'\n')
(root/'meter-resources.json').write_text(json.dumps({'exit_code':code, 'elapsed_seconds':time.time()-start,
    'sampled_peak_bytes':peak, 'sample_interval_seconds':1, 'before':before, 'after':snapshot()}, indent=2)+'\n')
