"""Run never-pushed scratch feature comparisons; preserve all phase logs."""
import json
import pathlib
import shutil
import subprocess
import sys

repository = pathlib.Path(sys.argv[1]).resolve()
adapted = pathlib.Path(sys.argv[2]).resolve()
scratch = pathlib.Path(sys.argv[3]).resolve()
scratch.mkdir(parents=True, exist_ok=True)
territory = pathlib.Path(__file__).resolve().parent
branch_prefix = sys.argv[4] if len(sys.argv) > 4 else 'scratch/latent-compare-'
features = [('main', None), ('taste', 'codex/taste-not-soundness'), ('flags', 'codex/flag-enums'), ('namespaces', 'codex/namespaces-tsc'), ('nested', 'codex/nested-functions')]
results = []

def run(command, cwd, log):
    with log.open('w') as output:
        return subprocess.run(command, cwd=cwd, stdout=output, stderr=subprocess.STDOUT).returncode

for name, feature in features:
    tree = scratch / name
    row = {'name': name, 'feature': feature, 'main': subprocess.check_output(['git', 'rev-parse', 'origin/main'], cwd=repository, text=True).strip()}
    results.append(row)
    branch = branch_prefix + name
    row['branch'] = branch
    row['status'] = 'worktree'
    code = run(['git', 'worktree', 'add', '-b', branch, str(tree), 'origin/main'], repository, scratch / (name + '-worktree.log'))
    if code == 0:
        row['status'] = 'merge'
        refs = ['origin/codex/stage3-base', 'origin/codex/tsc-census', 'origin/codex/stage3-type-imports', 'origin/codex/stage3-optional-declarations']
        if feature:
            refs.append('origin/' + feature)
            row['feature_sha'] = subprocess.check_output(['git', 'rev-parse', 'origin/' + feature], cwd=repository, text=True).strip()
        code = run(['git', 'merge', '--no-edit', *refs], tree, scratch / (name + '-merge.log'))
    if code == 0:
        row['head'] = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=tree, text=True).strip()
        # Share the already initialized read-only compiler dependency checkout.
        (tree / 'cohere').rmdir()
        (tree / 'cohere').symlink_to(repository / 'cohere', target_is_directory=True)
        shutil.copytree(territory, tree / 'stage3/census/latent')
        row['status'] = 'overlay'
        code = run(['python3', str(territory / 'make_overlay.py'), str(tree), str(scratch / (name + '-overlay'))], tree, scratch / (name + '-overlay.log'))
    if code == 0:
        row['status'] = 'build'
        code = run(['go', 'build', '-buildvcs=false', '-overlay=' + str(scratch / (name + '-overlay/overlay.json')), '-o', str(scratch / (name + '-census')), './stage3/census/latent/tool'], tree, scratch / (name + '-build.log'))
    if code == 0:
        row['status'] = 'run'
        code = run([str(scratch / (name + '-census')), str(adapted / 'src/compiler'), str(scratch / (name + '.jsonl'))], tree, scratch / (name + '-run.log'))
    if code == 0:
        row['status'] = 'complete'
    row['exit'] = code
    (scratch / 'runs.json').write_text(json.dumps(results, indent=2) + '\n')
    print(name, row['status'], code, flush=True)
