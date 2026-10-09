from pathlib import Path
import subprocess, os
root=Path('/workspace/adamic')
out=Path('/workspace/scratch/hidden-representation-mutants'); out.mkdir(exist_ok=True)
paths=[root/'docs/overload-results/run-mutants.py']+[root/f'docs/overload-results/groups/{group}/run-mutants.py' for group in ['callback','structural','fields','admission','field-hatch','values','visitors']]
for path in paths:
 group=path.parent.name if path.parent.name!='overload-results' else 'base'
 target=out/group;target.mkdir(exist_ok=True)
 script=path.read_text()
 # Keep __file__ for the repository root; redirect all witness output into scratch.
 script=script.replace("evidence = Path(__file__).parent / 'evidence'",f'evidence = Path({str(target)!r})')
 script=script.replace('out=Path(__file__).parent',f'out=Path({str(target)!r})')
 script=script.replace("Path(__file__).parent/(name+'-mutant.log.txt')",f"Path({str(target)!r})/(name+'-mutant.log.txt')")
 script=script.replace("(Path(__file__).parent/'mutants.json')",f"(Path({str(target)!r})/'mutants.json')")
 script=script.replace("'-timeout', '10m'", "'-timeout', '90s'")
 if "'-timeout'" not in script: script=script.replace("'-count=1'", "'-timeout','90s','-count=1'")
 temp=out/(group+'.py');temp.write_text(script)
 runner='exec(compile(open('+repr(str(temp))+').read(), '+repr(str(path))+', "exec"), {"__file__":'+repr(str(path))+', "__name__":"__main__"})'
 with (out/(group+'.log')).open('w') as log:
  result=subprocess.run(['python3','-c',runner],cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
 print(group,result.returncode,flush=True)
 if result.returncode: raise SystemExit(result.returncode)
