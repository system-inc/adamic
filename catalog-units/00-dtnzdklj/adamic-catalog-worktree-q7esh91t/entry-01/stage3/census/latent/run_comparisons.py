"""Measure resolved, never-pushed scratch worktrees. No output compiler is built."""
import concurrent.futures, json, os, pathlib, shutil, subprocess, sys, time
repository = pathlib.Path(sys.argv[1]).resolve()
adapted = pathlib.Path(sys.argv[2]).resolve()
scratch = pathlib.Path(sys.argv[3]).resolve(); scratch.mkdir(parents=True, exist_ok=True)
territory = pathlib.Path(__file__).resolve().parent
worktrees = json.loads(pathlib.Path(sys.argv[4]).read_text())
results = []
label = 'measured on a checker-rejected program'
def run(command, cwd, log, env=None):
    started = time.monotonic()
    with log.open('w') as output:
        code = subprocess.run(command, cwd=cwd, stdout=output, stderr=subprocess.STDOUT, env=env).returncode
    return code, round(time.monotonic()-started, 3)
def measure(spec):
    name = spec['name']; tree = pathlib.Path(spec['tree']).resolve()
    row = dict(spec, measurement=label, main=spec.get('main','ef3d907ecdc4c771b016f7d9c52372def057a340'))
    row['head'] = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=tree, text=True).strip()
    row['branch'] = subprocess.check_output(['git', 'branch', '--show-current'], cwd=tree, text=True).strip()
    source = pathlib.Path(spec.get('adapted', str(adapted))).resolve()
    row['adapted'] = str(source)
    if spec.get('unmerged'):
        assert row['head'] == row['main'] and not spec['features'], 'unmerged configuration changed'
    assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=repository/'cohere',text=True).strip() == subprocess.check_output(['git','rev-parse','HEAD:cohere'],cwd=tree,text=True).strip(), 'dependency pin mismatch'
    row['features'] = {feature: spec.get('pinned_commits',{}).get('origin/'+feature) or subprocess.check_output(['git', 'rev-parse', 'origin/'+feature], cwd=repository, text=True).strip() for feature in spec['features']}
    for sha in [row['main'], *row['features'].values()]:
        subprocess.run(['git', 'merge-base', '--is-ancestor', sha, row['head']], cwd=tree, check=True)
    dep = tree/'cohere'
    if not dep.is_symlink():
        dep.rmdir(); dep.symlink_to(repository/'cohere', target_is_directory=True)
    shutil.copytree(territory, tree/'stage3/census/latent', dirs_exist_ok=True)
    row['durations_seconds']={}
    # Use canonical dependency paths so independent worktrees share the warm Go cache.
    module = json.loads(subprocess.check_output(['go','mod','edit','-json'],cwd=tree,text=True))
    work = scratch/(name+'.go.work')
    replacements = [r for r in module['Replace'] if r['New']['Path'].startswith('./cohere/')]
    work.write_text('go 1.27\nuse (\n'+str(tree)+'\n'+str(repository/'cohere/TypeScript/tsc')+'\n)\nreplace (\n'+''.join(r['Old']['Path']+' => '+str(repository/r['New']['Path'])+'\n' for r in replacements)+')\n')
    environment = dict(os.environ,GOWORK=str(work),LATENT_ASSERT_NO_OUTPUT='1')
    phases = [('overlay' , ['python3', str(territory/'make_overlay.py'), str(tree), str(scratch/(name+'-overlay'))]),
              ('build', ['go', 'build', '-buildvcs=false', '-overlay='+str(scratch/(name+'-overlay/overlay.json')), '-o', str(scratch/(name+'-census')), './stage3/census/latent/tool']),
              ('run', [str(scratch/(name+'-census')), str(source/'src/compiler'), str(scratch/(name+'.jsonl'))])]
    for phase, command in phases:
        row['status'] = phase
        code, seconds = run(command, tree, scratch/(name+'-'+phase+'.log'), environment)
        row['durations_seconds'][phase]=seconds; row['exit']=code
        if code: break
    if not code: row['status']='complete'
    print(name, row['status'], code, row['durations_seconds'], flush=True)

    return row
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    pending = {pool.submit(measure,spec):spec for spec in worktrees}
    for future in concurrent.futures.as_completed(pending):
        results.append(future.result())
        order={r['name']:i for i,r in enumerate(worktrees)}
        results.sort(key=lambda r:order[r['name']])
        (scratch/'runs.json').write_text(json.dumps(results,indent=2)+'\n')
