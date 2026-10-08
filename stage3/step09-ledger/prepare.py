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
    overlap=[]
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
        if destination.name=='43-any-returns':
            rules_path=destination/'rules.json'
            rules=json.loads(rules_path.read_text())
            timer=[r for r in rules['rules'] if r['name']=='setTimeout']
            assert len(timer)==1
            main_rules=json.loads((ROOT/'stage3/adapt/41-explicit-any-remaining/rules.json').read_text())
            owner=[r for r in main_rules if r['before']==timer[0]['before']]
            assert len(owner)==1 and owner[0]['id']==13
            overlap.append(dict(stock=timer[0]['before'],retained=owner[0]['after'],skipped=timer[0]['after'],owner='41-explicit-any-remaining',skipped_rule='43-any-returns:setTimeout',reason='same stock return site already rewritten by main'))
            rules['rules']=[r for r in rules['rules'] if r['name']!='setTimeout']
            rules_path.write_text(json.dumps(rules,indent=2)+'\n')
        directories[destination.name]=destination
        extras.append(dict(adaptation=destination.name,sha=sha,source=territory))
    if '43-any-returns' in directories:
        # Preserve 71's parameter edits while matching 43's earlier return contracts.
        destination=out/'extra/71-writable-views'
        shutil.copytree(ROOT/'stage3/adapt/71-writable-views',destination,ignore=shutil.ignore_patterns('evidence'),dirs_exist_ok=True)
        sibling=out/'extra/70-readonly-views'
        if not sibling.exists():sibling.symlink_to(ROOT/'stage3/adapt/70-readonly-views',target_is_directory=True)
        spec_path=destination/'diagnostics.json'
        spec=json.loads(spec_path.read_text())
        for function,result in [('convertConfigFileToObject','AdamicJsonRecoveryObject'),('convertToJson','AdamicJsonRecoveryValue | undefined')]:
            selected=[edit for edit in spec['edits'] if ('function '+function+'(') in edit['before']]
            assert len(selected)==1
            edit=selected[0];original=dict(edit)
            for key in ['before','after']:
                assert edit[key].endswith('): any ')
                edit[key]=edit[key].removesuffix('any ')+result+' '
            overlap.append(dict(owner='43-any-returns',composition='71-writable-views',function=function,original=original,composed=edit,reason='71 changes parameter type; retain earlier 43 return type'))
        spec_path.write_text(json.dumps(spec,indent=2)+'\n')
        directories[destination.name]=destination
    transitions=[]
    snapshot_files={}
    for name,directory in sorted(directories.items()):
        before=out/(name+'-before');after=out/(name+'-after')
        for snapshot in [before,after]:
            snapshot.mkdir()
            if snapshot==after:
                if name=='43-any-returns':
                    text=(adapted/'src/compiler/sys.ts').read_text()
                    assert text.count(overlap[0]['retained'])==1 and overlap[0]['stock'] not in text
                subprocess.run(['node',str(directory/'adapt.cjs'),str(adapted)],check=True)
            for source in ['compiler','tsc']:
                for file in (adapted/'src'/source).rglob('*'):
                    if not file.is_file():continue
                    target=snapshot/source/file.relative_to(adapted/'src'/source)
                    target.parent.mkdir(exist_ok=True,parents=True)
                    digest=hashlib.sha256(file.read_bytes()).hexdigest()
                    if digest in snapshot_files:os.link(snapshot_files[digest],target)
                    else:
                        shutil.copyfile(file,target)
                        snapshot_files[digest]=target
        transitions.append(dict(adaptation=name,before=str(before),after=str(after)))
        if name.startswith(('40-','41-','42-','43-')):
            for version,snapshot in [('before',before),('after',after)]:
                shutil.copyfile(stock/'src/tsconfig-base.json',snapshot/'tsconfig-base.json')
                wrapper=out/(name+'-api-'+version);wrapper.mkdir()
                (wrapper/'src').symlink_to(snapshot);(wrapper/'node_modules').symlink_to(dependencies)
                inventory=out/(name+'-api-'+version+'.json')
                subprocess.run(['node',str(ROOT/'stage3/step09-ledger/enumerate.cjs'),str(wrapper),str(inventory)],check=True)
                transitions[-1]['semantic_'+version]=str(inventory)
    (out/'composition.json').write_text(json.dumps(overlap,indent=2)+'\n')
    (out/'transitions.json').write_text(json.dumps(transitions,indent=2)+'\n')
    provenance=dict(repository_sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),source=pin,extras=extras,node_types=node_types['version'],lock_sha256=hashlib.sha256((stock/'package-lock.json').read_bytes()).hexdigest())
    (out/'PROVENANCE.json').write_text(json.dumps(provenance,indent=2)+'\n')


if __name__=='__main__':main()
