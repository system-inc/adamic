import importlib.util,sys,json,subprocess,os,time,inspect
from pathlib import Path
sys.dont_write_bytecode=True
repo=Path('/workspace/adamic');base=Path('/workspace/scratch/scanner-main-next-records');out=base/'supplemental-generic-value';out.mkdir();(out/'split-0').symlink_to(base/'split-0',target_is_directory=True)
spec=importlib.util.spec_from_file_location('stops',repo/'stage3/drivers/scanner/scratch-stops.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
original=m.catalogue
m.catalogue=lambda:original()+[{'source':'export const joined: string = "token:" + 7; console.log(joined);\n','message':"stage 0 can't lower a BinaryExpression with a string and a number yet",'provenance':'fresh mixed-string-number witness'}, {'source':'const {length} = "token"; console.log(length.toString());\n','message':"stage 0 can't lower destructuring a string yet",'provenance':'fresh string-destructuring witness'}, {'source':'let index = 0; const previous = index++; console.log(previous.toString());\n','message':"stage 0 can't lower a PostfixUnaryExpression yet",'provenance':'fresh postfix-value witness'}, {'source':'function id<T>(value: T): T { return value; } const alias = id; console.log(alias("ok"));\n','message':"stage 0 can't lower a generic function as a value yet",'provenance':'fresh generic-function-value witness'}]
# Continue the existing placeholder state for at most eight additional stops.
source=inspect.getsource(m.walk).replace('range(1, 16)','range(1, 9)').replace('if order == 15:','if order == 8:')
source=source.replace('range(1, 9)', 'range(1, 5)').replace('if order == 8:', 'if order == 4:')
exec(source,m.__dict__)
phases=[]
def run(command,name,cwd,extra=None):
 begin=time.monotonic()
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
  code=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; exec "$@"','tools',*map(str,command)],cwd=cwd,env=dict(os.environ,**(extra or {})),stdout=stdout,stderr=stderr).returncode
 phases.append({'name':name,'command':list(map(str,command)),'exit':code,'wall_seconds':round(time.monotonic()-begin,3)})
 (out/'phases.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,code,flush=True);return code
rows=m.walk(out,base/'supplemental-postfix/discovery/adapted',base/'adamic',base/'compiler-tree',Path('/workspace/scratch/native3-cache'),run)
for row in rows:
 row['order']+=11;row['scope']='behind the original command and supplementary throwing placeholders'
 if row.get('continuation')=='fifteen-stop bound reached':row['continuation']='fifteen total stops reached'
(out/'combined-stops.json').write_text(json.dumps(json.loads((base/'supplemental-postfix/combined-stops.json').read_text())[:11]+rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
