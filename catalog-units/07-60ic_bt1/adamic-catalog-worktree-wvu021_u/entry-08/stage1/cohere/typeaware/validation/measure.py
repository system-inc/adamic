import os, re, subprocess, time, statistics, pathlib, sys
root = pathlib.Path(sys.argv[1]).resolve()
repo = pathlib.Path(__file__).resolve().parents[4]
config = repo / 'stage1/cohere/typeaware/testdata/tsconfig.json'
env = dict(os.environ, ADAMIC_TSGO_TIMING='1')
def run(name, args):
    out = root / (name + '.stdout')
    err = root / (name + '.stderr')
    with out.open('wb') as o, err.open('wb') as e:
        start = time.perf_counter()
        subprocess.run(list(map(str, args)), cwd=repo, env=env, stdout=o, stderr=e, check=True)
        elapsed = time.perf_counter() - start
    report = err.read_text().strip()
    print(name, f'{elapsed:.9f}s', out.read_text().strip(), report, flush=True)
    return out.read_bytes(), elapsed, report
print('machine load', pathlib.Path('/proc/loadavg').read_text().strip(), flush=True)
print('CPU', next(x for x in pathlib.Path('/proc/cpuinfo').read_text().splitlines() if x.startswith('model name')), flush=True)
costs = {'native': [], 'Go': []}
for round in range(1, 4):
    outputs = []
    for side in (['native', 'Go'] if round % 2 else ['Go', 'native']):
        args = [root/'query-cost', config, root/'cost-probe.ts', '10000'] if side == 'native' else [root/'oracle', config, root/'probe.manifest', '--query-cost', '10000']
        output, elapsed, report = run(f'isolated-query-{round}-{side}', args)
        outputs.append(output)
        fields = {k: int(v) for k,v in re.findall(r'(query_ns|queries|first_query_ns)=(\d+)', report)}
        warm = (fields['query_ns'] - fields['first_query_ns']) / (fields['queries'] - 1)
        costs[side].append(warm)
        print(f'warm {side} {warm:.3f} ns/query', flush=True)
    assert outputs[0] == outputs[1] == b'90000\n'
for side, values in costs.items():
    print(f'median warm {side} {statistics.median(values):.3f} ns/query', flush=True)
for corpus, configPath, manifest in [
    ('compiler', pathlib.Path(sys.argv[2]).resolve() / 'src/compiler/tsconfig.json', root/'compiler.manifest'),
    ('generated', config, root/'generated.manifest'),
]:
    times = {'native': [], 'Go': []}
    for round in range(1,4):
        outputs = []
        for side in (['Go', 'native'] if round%2 else ['native', 'Go']):
            binary = root/('native' if side=='native' else 'oracle')
            output, elapsed, report = run(f'isolated-{corpus}-{round}-{side}', [binary, configPath, manifest, '--count'])
            outputs.append(output)
            times[side].append(elapsed)
        assert outputs[0] == outputs[1]
    count = int(re.search(rb'findings (\d+)', outputs[0]).group(1))
    for side, values in times.items():
        median = statistics.median(values)
        print(f'median {corpus} {side} {median:.9f}s {count/median:.3f} findings/s', flush=True)
print('load after', pathlib.Path('/proc/loadavg').read_text().strip())
