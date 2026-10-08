"""Materialize only pinned typed sources; no installs, hooks, or repo scripts."""
import concurrent.futures, json, os, pathlib, subprocess, sys
pins = json.loads(pathlib.Path(sys.argv[1]).read_text())
def run(p):
    base = ['git', '-C', p['root'], '-c', 'core.hooksPath=/dev/null']
    def git(*args, **kw):
        return subprocess.run(base + list(args), check=True, capture_output=True, **kw)
    sha = git('rev-parse', 'FETCH_HEAD').stdout.decode().strip()
    if sha != p['sha']: raise RuntimeError('pin drift: '+p['repo'])
    tree = git('ls-tree', '-rz', 'FETCH_HEAD').stdout.split(b'\0')
    entries = [e.split(b'\t', 1) for e in tree if e and e.split(b'\t',1)[1].endswith((b'.ts',b'.tsx'))]
    ids = sorted({head.split()[2].decode() for head, _ in entries})
    env = dict(os.environ, GIT_NO_LAZY_FETCH='1')
    checks = git('cat-file', '--batch-check=%(objectname) %(objecttype)', input=('\n'.join(ids)+'\n').encode(), env=env).stdout.decode().splitlines()
    missing = [x.split()[0] for x in checks if x.endswith(' missing')]
    for index in range(0,len(missing),512):
        git('fetch','--no-tags','--no-write-fetch-head','--recurse-submodules=no','--stdin','https://github.com/'+p['repo']+'.git',input=('\n'.join(missing[index:index+512])+'\n').encode())
    git('--literal-pathspecs','restore','--source=FETCH_HEAD','--worktree','--pathspec-from-file=-','--pathspec-file-nul', input=b'\0'.join(path for _,path in entries)+b'\0')
    print(p['repo'],p['sha'],len(entries),'typed files',flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
    list(executor.map(run,pins))
