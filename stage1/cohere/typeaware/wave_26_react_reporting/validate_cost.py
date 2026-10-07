"""Three alternating quiet process timings; every run compares full diagnostic bytes."""
import argparse
import json
import pathlib
import statistics
import subprocess
import time
ROOT = pathlib.Path(__file__).resolve().parents[4]
p = argparse.ArgumentParser()
p.add_argument('--scratch', required=True)
p.add_argument('--static', required=True)
p.add_argument('--render', required=True)
p.add_argument('--effect', required=True)
p.add_argument('--compiler-manifest', required=True)
p.add_argument('--repository-manifest', required=True)
a = p.parse_args()
scratch = pathlib.Path(a.scratch).resolve()
scratch.mkdir(parents=True, exist_ok=True)
records = []
for rule in ['static', 'render', 'effect']:
    directory = pathlib.Path(getattr(a, rule)).resolve()
    for population in ['compiler', 'repository']:
        values = {'go': [], 'native': []}
        expected = (directory / (population + '-go.stdout')).read_bytes()
        for round in range(3):
            for engine in (['go', 'native'] if round % 2 == 0 else ['native', 'go']):
                name = f'{rule}-{population}-{round}-{engine}'
                command = [str(directory / ('oracle' if engine == 'go' else 'native')), str(directory / 'tsconfig.json'), getattr(a, population + '_manifest')]
                started = time.monotonic()
                with (scratch / (name + '.stdout')).open('wb') as out, (scratch / (name + '.stderr')).open('wb') as err:
                    result = subprocess.run(command, cwd=ROOT, stdout=out, stderr=err)
                elapsed = time.monotonic() - started
                assert result.returncode == 0
                assert (scratch / (name + '.stdout')).read_bytes() == expected
                if engine == 'native':
                    assert (scratch / (name + '.stderr')).read_bytes() == b''
                values[engine].append(elapsed)
                records.append(dict(name=name, command=command, exit=result.returncode, seconds=elapsed))
                (scratch / 'runs.json').write_text(json.dumps(records, indent=2) + '\n')
        print(rule, population, 'native', f'{statistics.median(values["native"]):.6f}', 'Go', f'{statistics.median(values["go"]):.6f}', flush=True)
