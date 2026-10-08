"""Prove output guards using a file entry, as the frozen entry census requires."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys


def main(repository, baseline_overlay, baseline_binary, output):
    repository, baseline_overlay, baseline_binary, output = map(lambda value: Path(value).resolve(), (repository, baseline_overlay, baseline_binary, output))
    output.mkdir(exist_ok=True)
    source = output / 'bad.a'
    source.write_text('function wrong(): number { return "wrong"; }\n')
    environment = dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1')
    with (output / 'baseline.log.txt').open('w') as log:
        subprocess.run([str(baseline_binary), str(source), str(output / 'baseline.jsonl')], env=environment, stdout=log, stderr=subprocess.STDOUT, check=True)
    print('PASS: baseline full entry census accepts the witness for measurement with both guards enabled', flush=True)
    for name, filename, old, new, expected in [
        ('IR', 'internal_lower_lower.go', 'return nil, fmt.Errorf("latent census: measurement only; no IR output")', 'return &ir.Program{}, fmt.Errorf("latent census: measurement only; no IR output")', 'measurement returned usable IR'),
        ('loader', 'internal_load_load.go', 'return nil, fmt.Errorf("latent census: measurement loader only; no output path")', 'return LatentLoad(paths)', 'measurement loader exposed an output program'),
    ]:
        folder = output / name
        shutil.copytree(baseline_overlay, folder, dirs_exist_ok=True)
        path = folder / filename
        text = path.read_text()
        assert text.count(old) == 1
        path.write_text(text.replace(old, new, 1))
        original = json.loads((baseline_overlay / 'overlay.json').read_text())
        overlay = {'Replace': {name: str(folder / Path(value).name) for name, value in original['Replace'].items()}}
        (folder / 'overlay.json').write_text(json.dumps(overlay, indent=2) + '\n')
        binary = folder / 'census'
        with (folder / 'build.log.txt').open('w') as log:
            subprocess.run(['go', 'build', '-buildvcs=false', '-overlay=' + str(folder / 'overlay.json'), '-o', str(binary), './stage3/census/latent/tool'], cwd=repository, stdout=log, stderr=subprocess.STDOUT, check=True)
        with (folder / 'guard.log.txt').open('w') as log:
            result = subprocess.run([str(binary), str(source), str(folder / 'mutant.jsonl')], env=environment, stdout=log, stderr=subprocess.STDOUT)
        assert result.returncode != 0 and expected in (folder / 'guard.log.txt').read_text(), (name, result.returncode)
        print(name + ' guard mutant caught: ' + expected, flush=True)


if __name__ == '__main__':
    main(*sys.argv[1:])
