from pathlib import Path
import sys,os,json,hashlib
sys.path.insert(0,'/workspace/scratch/sys-walls');from run import run,S,ROOT
out=Path('/workspace/sys-recordreplay-report/stage3/census/sys-recordreplay')
def checked(name,args,cwd=ROOT):
 code=run(name,args,cwd)
 if code:raise SystemExit(code)
tree=S/'recordreplay';checked('recordreplay-copy',['cp','-a',S/'before',tree]);modules=tree/'node_modules'
if not modules.exists():modules.symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
checked('recordreplay-instrument',['node',S/'prepare-recordreplay.cjs',tree])
base=S/'before/src';names=sorted(p.relative_to(base) for p in base.rglob('*') if p.is_file());changed=[str(n) for n in names if (base/n).read_bytes()!=(tree/'src'/n).read_bytes()];assert changed==['compiler/sys.ts'],changed
(out/'source-scope.json').write_text(json.dumps({'files':len(names),'changed':changed,'old':'!!(process as any).recordreplay','new':'!!(globalThis as any).__sysRecordReplay((process as any).recordreplay, process)'},indent=2)+'\n')
checked('recordreplay-build',['npm','run','build'],tree)
checked('recordreplay-harness',['npm','exec','hereby','--','tests'],tree)
modules.unlink()
owner=(S/'recordreplay-owner.ts.txt').read_text();report=(out/'REPORT.md').read_text().replace('The enclosing function will be quoted here from the completed full apply before instrumentation.','The enclosing function below is quoted whole from the completed full apply before instrumentation.\n\n```ts\n'+owner+'\n```');(out/'REPORT.md').write_text(report)
print('RECORDREPLAY PREPARED',flush=True)
