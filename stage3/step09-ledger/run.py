"""Run API inventories, production cast proof, reconciliation and fixture over prepared trees."""
import argparse
import json
import os
from pathlib import Path
import subprocess

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[1]


def main():
    p=argparse.ArgumentParser();p.add_argument('prepared',type=Path);p.add_argument('output',type=Path);args=p.parse_args()
    prepared=args.prepared.resolve();out=args.output.resolve();out.mkdir()
    def run(command,name,env=None):
        with (out/name).open('w') as log:
            subprocess.run(command,cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True,env=env)
    for version in ['stock','adapted']:
        run(['node',str(HERE/'enumerate.cjs'),str(prepared/version),str(out/(version+'.json'))],version+'-inventory.txt')
    overlay=out/'overlay'
    run(['python3',str(HERE/'make_overlay.py'),str(overlay)],'overlay.txt')
    probe=out/'cast-probe'
    run(['go','build','-overlay',str(overlay/'overlay.json'),'-o',str(probe),'./stage3/step09-ledger/probe'],'probe-build.txt')
    adapted=json.loads((out/'adapted.json').read_text())
    roots=[str(prepared/'adapted'/f['file']) for f in adapted['files']]
    roots.append(str(prepared/'adapted/node_modules/@types/node/index.d.ts'))
    with (out/'casts.json').open('w') as result,(out/'casts-errors.txt').open('w') as error:
        subprocess.run([str(probe),*roots],cwd=ROOT,stdout=result,stderr=error,check=True)
    run(['python3',str(HERE/'ledger.py'),str(out/'stock.json'),str(out/'adapted.json'),str(prepared/'transitions.json'),str(out/'casts.json'),str(out/'result')],'reconcile.txt')
    run(['node',str(HERE/'test_overlap.cjs'),str(prepared),str(out/'result')],'timer-overlap.txt')
    run(['python3',str(HERE/'test_fixture.py')],'fixture.txt',dict(os.environ,STEP09_CAST_PROBE=str(probe)))
    run(['go','vet','-overlay',str(overlay/'overlay.json'),'./stage3/step09-ledger/probe'],'probe-vet.txt')
    print(json.dumps(json.loads((out/'result/SUMMARY.json').read_text())['dispositions'],indent=2))


if __name__=='__main__':main()
