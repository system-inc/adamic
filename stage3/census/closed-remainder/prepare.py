#!/usr/bin/env python3
"""Prepare symmetric scoped-tail overlays from each measured compiler's stock harness."""
import json,pathlib,subprocess,sys,os,signal
repo,out=map(lambda s:pathlib.Path(s).resolve(),sys.argv[1:]); here=pathlib.Path(__file__).resolve().parent
def run(*args):
 argv=[str(a) for a in args]; child=subprocess.run(argv,cwd=repo)
 print(json.dumps({'argv':argv,'child_exit_code':child.returncode}),flush=True)
 if child.returncode<0:
  if -child.returncode not in (signal.SIGKILL,signal.SIGSTOP):signal.signal(-child.returncode,signal.SIG_DFL)
  os.kill(os.getpid(),-child.returncode)
 if child.returncode:sys.exit(child.returncode)
run('python3',repo/'stage3/census/latent/make_overlay.py',repo,out/'stock')
run('python3',repo/'stage3/meter/entry_overlay.py',out/'stock',out/'entry')
run('python3',here/'prototype.py',out/'entry',out/'tail')
run('python3',here/'isolate-and-trace.py',out/'tail',out/'final',repo)
run('python3',here/'add-predicate-trace.py',repo,out/'final')
p=out/'final/hook.go';s=p.read_text();a='probe := *l\n if os.Getenv("LWP_ISOLATE_PROBE")=="1" {probe = *l.latentSnapshot()}';assert s.count(a)==1;p.write_text(s.replace(a,'probe := *l.latentSnapshot()'))
print(out/'final/overlay.json')
