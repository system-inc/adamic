#!/usr/bin/env python3
import gzip,json,shutil,subprocess,sys,tempfile
from pathlib import Path
here=Path(__file__).resolve().parent
subprocess.run([sys.executable,str(here/'verify.py')],check=True)
for name in ('pin','split','node','ranking','native-claim','population','owner','header','golden','request'):
    with tempfile.TemporaryDirectory(prefix='step31-proof-mutant-') as tmp:
        out=Path(tmp)/'proof';shutil.copytree(here,out)
        p=out/'result.json';d=json.loads(p.read_text())
        if name=='pin':d['candidate']='0'*40
        elif name=='split':d['profiles'][1]['builds'][1]['stderr']+='changed'
        elif name=='node':d['probe']['node']['stdout']='wrong\n'
        elif name=='ranking':d['profiles'][1]['ranking_matches']=[{'invented':True}]
        elif name=='native-claim':d['native_comparator_run']=True
        elif name=='population':d['node_manifest']['projects'].pop()
        elif name=='owner':d['profiles'][1]['owner']='compiler'
        elif name in ('golden','request'):
            file=out/'evidence'/('node-golden.stdout.gz' if name=='golden' else 'node-request.json.gz')
            file.write_bytes(gzip.compress(gzip.decompress(file.read_bytes())+b'changed',mtime=0))
        else:(out/'probes/commonjs.a').write_text((out/'probes/commonjs.a').read_text().split('\n',1)[1])
        p.write_text(json.dumps(d))
        r=subprocess.run([sys.executable,str(here/'verify.py'),str(out)],capture_output=True)
        assert r.returncode!=0 and b'AssertionError' in r.stderr,name
        (here/'evidence'/(name+'-mutant.txt')).write_bytes(r.stderr)
        print(name+': caught by verifier assertion')
print('10/10 evidence mutants caught')
