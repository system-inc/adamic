"""Repeat cold program + warm answer/cache and owner-pool comparisons."""
import json, os, pathlib, subprocess, sys, time
measure, native, config, manifest, output = sys.argv[1:]
root = pathlib.Path(output)
root.mkdir()
records = []
expected = None
for round in range(3):
    env = dict(os.environ, GOMAXPROCS='4', ADAMIC_TSGO_TIMING='1')
    out = root / f'experiment-{round}.json'
    subprocess.run([measure, 'experiment', config, manifest, str(out)], env=env, check=True)
    for cache in ([False, True] if round % 2 == 0 else [True, False]):
        command = [native, config, str(out)+'.requests', '1000'] + (['cache'] if cache else [])
        start = time.perf_counter_ns()
        result = subprocess.run(command, env=env, capture_output=True, check=True)
        elapsed = time.perf_counter_ns()-start
        if expected is None: expected = result.stdout
        if result.stdout != expected: raise RuntimeError('cache or round answer bytes drifted')
        records.append(dict(round=round, cache=cache, process_ns=elapsed, stderr=result.stderr.decode()))
(root/'native-cache.json').write_text(json.dumps(records, indent=2)+'\n')
