from pathlib import Path
import json,subprocess,os,shutil,time
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');tree=s/'counted'
shutil.copytree(s/'before',tree,symlinks=True)
p=tree/'src/compiler/core.ts';text=p.read_text();before='''export function arrayOf<T>(count: number, f: (index: number) => T): T[] {
    const result = new Array(count);
    for (let i = 0; i < count; i++) {
        result[i] = f(i);
    }
    return result;
}'''
assert text.count(before)==1
after=before.replace('    const result', '    (globalThis as any).__arrayOfBegin(count);\n    const shadow: T[] = [];\n    const result').replace('result[i] = f(i);','let value: T;\n        try { value = f(i); } catch (error) { (globalThis as any).__arrayOfThrow(count, i, error); throw error; }\n        result[i] = value;\n        shadow.push(value);').replace('    return result;','    (globalThis as any).__arrayOfEnd(count, result, shadow);\n    return result;')
p.write_text(text.replace(before,after));(s/'instrument.json').write_text(json.dumps(dict(before=before,after=after,measurementOnly=True),indent=2)+'\n')
env=dict(os.environ,NODE_OPTIONS='--require='+str(s/'preload.cjs'))
for name,command in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
 start=time.monotonic()
 with (s/(name+'-counted.log')).open('w') as log:r=subprocess.run(command,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1700)
 print(name,r.returncode,time.monotonic()-start,flush=True)
 assert r.returncode==0

(tree/"node_modules").rename(s/"counted-build-packages") # supplied-product shard views install their own node_modules link
