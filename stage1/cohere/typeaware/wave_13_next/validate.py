"""Compare complete production Go and native finding bytes; keep every process log."""
import hashlib, json, os, pathlib, statistics, subprocess, sys, time
native, oracle, controls, output = map(pathlib.Path, sys.argv[1:5])
output.mkdir(parents=True, exist_ok=True)
def run(name, binary, config, manifest, timed=False):
    env = dict(os.environ)
    if timed: env['ADAMIC_TSGO_TIMING'] = '1'
    out, err = output / (name + '.stdout'), output / (name + '.stderr')
    started = time.perf_counter()
    with out.open('wb') as stdout, err.open('wb') as stderr:
        result = subprocess.run([str(binary), str(config), str(manifest)], stdout=stdout, stderr=stderr, env=env)
    elapsed = time.perf_counter() - started
    return result.returncode, out.read_bytes(), err.read_bytes(), elapsed
results = []
failures = 0
for case in sorted(controls.glob('*/case.json')):
    info = json.loads(case.read_text())
    config, manifest = case.parent / 'tsconfig.json', case.parent / 'roots.manifest'
    want = run(case.parent.name + '-go', oracle, config, manifest)
    got = run(case.parent.name + '-native', native, config, manifest)
    okay = want[0] == 0 and got[0] == 0 and got[2] == b'' and want[1] == got[1]
    offset = next((at for at, (a,b) in enumerate(zip(want[1],got[1])) if a != b), min(len(want[1]),len(got[1])))
    results.append(dict(test=info['test'], case=case.parent.name, equal=okay, go_exit=want[0], native_exit=got[0], bytes=len(want[1]), findings=want[1].splitlines()[-1].decode() if want[1] else '', difference=offset, sha256=hashlib.sha256(want[1]).hexdigest()))
    if not okay:
        failures += 1
        print(info['test'], 'go',want[0],'native',got[0], 'difference',offset, got[2].decode(errors='replace')[:300], flush=True)
(output / 'controls-results.json').write_text(json.dumps(results,indent=2)+'\n')
print('cases',len(results),'failures',failures,flush=True)
raise SystemExit(bool(failures))
