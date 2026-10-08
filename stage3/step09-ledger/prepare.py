"""Prepare stock, adapted and incremental source snapshots without compiler merges."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

ROOT=Path(__file__).resolve().parents[2]


def main():
    p=argparse.ArgumentParser()
    p.add_argument('output',type=Path)
    p.add_argument('--node-modules',required=True,type=Path,help='upstream npm ci output, matching its lockfile')
    p.add_argument('--extra',action='append',default=[],help='git-ref:stage3/adapt/NN-name; copy this adapter into scratch only')
    args=p.parse_args();out=args.output.resolve()
    out.mkdir();stock=out/'stock';adapted=out/'adapted'
    pin=json.loads((ROOT/'stage3/source.json').read_text())
    cache=Path(os.environ.get('STAGE3_CACHE',str(Path.home()/'.cache/adamic-stage3')))
    subprocess.run(['git','clone','--no-checkout','--no-hardlinks',str(cache/'typescript.git'),str(stock)],check=True)
    subprocess.run(['git','-C',str(stock),'sparse-checkout','set','--no-cone','/src/','/scripts/','/package.json','/package-lock.json','/tests/baselines/reference/api/typescript.d.ts'],check=True)
    subprocess.run(['git','-C',str(stock),'checkout','--detach',pin['commit']],check=True)
    subprocess.run(['node',str(ROOT/'stage3/adapt/00-setup/adapt.cjs'),str(stock)],check=True)
    dependencies=args.node_modules.resolve()
    lock=json.loads((stock/'package-lock.json').read_text())
    node_types=json.loads((dependencies/'@types/node/package.json').read_text())
    assert node_types['version']==lock['packages']['node_modules/@types/node']['version']
    (stock/'node_modules').symlink_to(dependencies)
    # Only the source and build metadata used by this census, plus adaptation 40's API baseline.
    adapted.mkdir()
    for name in ['src','scripts','tests']:
        shutil.copytree(stock/name,adapted/name)
    for name in ['package.json','package-lock.json']:
        shutil.copyfile(stock/name,adapted/name)
    (adapted/'node_modules').symlink_to(dependencies)
    directories={d.name:d for d in (ROOT/'stage3/adapt').iterdir() if d.is_dir() and d.name!='00-setup'}
    extras=[]
    for spec in args.extra:
        ref,territory=spec.split(':',1)
        sha=subprocess.check_output(['git','rev-parse',ref],cwd=ROOT,text=True).strip()
        destination=out/'extra'/Path(territory).name
        names=subprocess.check_output(['git','ls-tree','-r','--name-only',sha,territory],cwd=ROOT,text=True).splitlines()
        assert names,('missing adaptation',spec)
        for name in names:
            if '/evidence/' in name:continue
            target=destination/Path(name).relative_to(territory)
            target.parent.mkdir(exist_ok=True,parents=True)
            target.write_bytes(subprocess.check_output(['git','show',sha+':'+name],cwd=ROOT))
        directories[destination.name]=destination
        extras.append(dict(adaptation=destination.name,sha=sha,source=territory))
    transitions=[]
    for name,directory in sorted(directories.items()):
        before=out/(name+'-before');after=out/(name+'-after')
        for snapshot in [before,after]:
            snapshot.mkdir()
            if snapshot==after:
                subprocess.run(['node',str(directory/'adapt.cjs'),str(adapted)],check=True)
            for source in ['compiler','tsc']:
                shutil.copytree(adapted/'src'/source,snapshot/source)
        transitions.append(dict(adaptation=name,before=str(before),after=str(after)))
        if name.startswith(('40-','41-','42-','43-')):
            for version,snapshot in [('before',before),('after',after)]:
                shutil.copyfile(stock/'src/tsconfig-base.json',snapshot/'tsconfig-base.json')
                wrapper=out/(name+'-api-'+version);wrapper.mkdir()
                (wrapper/'src').symlink_to(snapshot);(wrapper/'node_modules').symlink_to(dependencies)
                inventory=out/(name+'-api-'+version+'.json')
                subprocess.run(['node',str(ROOT/'stage3/step09-ledger/enumerate.cjs'),str(wrapper),str(inventory)],check=True)
                transitions[-1]['semantic_'+version]=str(inventory)
    (out/'transitions.json').write_text(json.dumps(transitions,indent=2)+'\n')
    provenance=dict(repository_sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),source=pin,extras=extras,node_types=node_types['version'],lock_sha256=hashlib.sha256((stock/'package-lock.json').read_bytes()).hexdigest())
    (out/'PROVENANCE.json').write_text(json.dumps(provenance,indent=2)+'\n')


if __name__=='__main__':main()
