import subprocess, pathlib, time
out=pathlib.Path('review/test-audit/internal-load-pin')
repo='cohere/TypeScript'
def git(*args): return subprocess.check_output(['git','-C',repo,*args],text=True).strip()
pin=git('rev-parse','HEAD'); tree=git('rev-parse','HEAD^{tree}')
constructed=subprocess.check_output(['git','-C',repo,'-c','user.name=Test audit','-c','user.email=test-audit@example.invalid','commit-tree',tree,'-p',pin],input='Scratch setup probe: same source tree with a different commit identity.\n',text=True).strip()
start=time.monotonic()
try:
 subprocess.run(['git','-C',repo,'checkout','--detach',constructed],check=True,capture_output=True)
 with (out/'S1.log').open('w') as log:
  result=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','^TestTypeScriptIsThePinnedCommit$'],stdout=log,stderr=subprocess.STDOUT)
 (out/'S1.txt').write_text(f'Original HEAD: {pin}\nConstructed HEAD: {constructed}\nIdentical tree: {tree}\nExit: {result.returncode}\nWall seconds: {time.monotonic()-start:.3f}\nNo test or git oracle edited. The repository checkout construction was changed. No production mutant kill claimed.\n')
finally: subprocess.run(['git','-C',repo,'checkout','--detach',pin],check=True,capture_output=True)
