#!/usr/bin/env python3
"""Integrate each requested feature alone and run the actual stage-0 gate."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
from merge import index_conflicts

repo = Path(__file__).resolve().parents[3]
scratch = Path(sys.argv[1]).resolve()
tree = Path(sys.argv[2]).resolve()
scratch.mkdir(parents=True, exist_ok=True)
profiles = [('main', []), ('taste', ['taste-not-soundness']), ('flags', ['flag-enums']),
            ('namespaces', ['namespaces-tsc']), ('nested', ['nested-functions']),
            ('combined', ['taste-not-soundness', 'flag-enums', 'namespaces-tsc', 'nested-functions'])]
base = 'ef3d907ecdc4c771b016f7d9c52372def057a340'
for name, features in profiles:
    work = scratch / name
    if work.exists():
        raise SystemExit(f'refusing existing worktree: {work}')
    log = scratch / (name + '.log')
    record = {'base': base, 'features': [], 'conflicts': [], 'entryResults': {}}
    with log.open('w') as output:
        def run(*args, cwd=work, check=True):
            return subprocess.run(args, cwd=cwd, stdout=output, stderr=subprocess.STDOUT, check=check)
        run('git', 'worktree', 'add', '-b', 'scratch/parser-' + name, str(work), base, cwd=repo)
        for feature in features:
            ref = 'origin/codex/' + feature
            sha = subprocess.check_output(['git', 'rev-parse', ref], cwd=repo, text=True).strip()
            record['features'].append({'branch': ref, 'sha': sha})
            result = run('git', 'merge', '--no-edit', ref, check=False)
            if result.returncode:
                conflicts = subprocess.check_output(['git', 'diff', '--name-only', '--diff-filter=U'], cwd=work, text=True).splitlines()
                if not conflicts: raise RuntimeError(f'merge failed without resolvable conflicts: {name}')
                record['conflicts'].append({'feature': feature, 'files': conflicts, 'resolution': 'three-way hunk reconstruction with merge.py; retain main and feature changes'})
                index_conflicts(work, conflicts)
                run('gofmt', '-w', *[file for file in conflicts if file.endswith('.go')])
                run('git', 'add', '--', *conflicts)
                run('git', 'commit', '-m', 'Resolve scratch feature conflicts retaining both changes')
        # Reuse the initialized, pinned cohere checkout rather than fetching it again.
        cohere = work / 'cohere'
        if cohere.is_dir() and not any(cohere.iterdir()): cohere.rmdir()
        if not cohere.exists(): cohere.symlink_to(repo / 'cohere', target_is_directory=True)
        go_mod = work / 'go.mod'
        go_mod.write_text(go_mod.read_text().replace('=> ./cohere/', '=> ' + str(repo / 'cohere') + '/'))
        record['submoduleInputs'] = 'same pinned cohere checkout via absolute module replacement paths'
        probe = work / 'stage3/drivers/parser/probe'
        probe.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(repo / 'stage3/drivers/parser/probe/main.go', probe / 'main.go')
        source = (work / 'internal/load/load.go').read_text()
        # External input integration only: Node declarations replace the Adamic
        # prelude for upstream compiler source. All soundness/subset options remain.
        assert source.count('roots = append(roots, preludePath)') == 1
        source = source.replace('roots = append(roots, preludePath)', '// Upstream source uses pinned Node declarations, not the Adamic prelude.')
        assert source.count('roots[:len(roots)-1]') == 1
        source = source.replace('roots[:len(roots)-1]', 'roots')
        assert source.count('Types:            []string{},') == 1
        source = source.replace('Types:            []string{},', 'Types:            []string{"node"},')
        assert 'NoImplicitReturns:          core.TSTrue,' in source
        assert 'NoFallthroughCasesInSwitch: core.TSTrue,' in source
        overlay_source = scratch / (name + '-load.go.txt')
        overlay_source.write_text(source)
        overlay = scratch / (name + '-overlay.json')
        overlay.write_text(json.dumps({'Replace': {str(work / 'internal/load/load.go'): str(overlay_source)}}))
        binary = scratch / (name + '-probe')
        built = run('go', 'build', '-buildvcs=false', '-overlay=' + str(overlay), '-o', str(binary), './stage3/drivers/parser/probe', check=False)
        record['buildExit'] = built.returncode
        if built.returncode == 0:
            for entry in ['parser.ts', 'scanner.ts']:
                dest = scratch / (name + '-' + entry + '.json')
                result = run(str(binary), str(tree / 'src/compiler' / entry), str(dest), cwd=tree, check=False)
                record['entryResults'][entry] = {'exit': result.returncode, 'result': str(dest)}
        (scratch / (name + '-integration.json')).write_text(json.dumps(record, indent=2) + '\n')
    print(name, 'build', record['buildExit'], record['entryResults'], flush=True)
