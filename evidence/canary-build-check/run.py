import json, pathlib, shlex, subprocess
root=pathlib.Path('/workspace/adamic/evidence/canary-build-check')
work=pathlib.Path('/workspace/adamic-rule-g-2594029')
restore='e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1'
results={}
def run(label,args):
    args=list(map(str,args))
    (root/(label+'.command')).write_text(shlex.join(args)+'\n')
    r=subprocess.run(args,cwd=work,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    (root/(label+'.stdout')).write_bytes(r.stdout)
    (root/(label+'.stderr')).write_bytes(r.stderr)
    (root/(label+'.exit')).write_text(str(r.returncode)+'\n')
    results[label]={'exit':r.returncode,'stdout':r.stdout.decode(errors='replace'),'stderr':r.stderr.decode(errors='replace')}
    (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
    return r.returncode
run('01-go-version',['go','version'])
try:
    for short,sha in [('22b8a0b','22b8a0b9aa55c785165d8ea75556d853cacd8602'),('ad1ddfe','ad1ddfe1295cbdea512db15108358c75e522426a')]:
        if run(short+'-checkout',['git','switch','--detach',sha]): raise RuntimeError('checkout failed')
        run(short+'-revision',['git','rev-parse','HEAD'])
        run(short+'-submodules',['git','submodule','status','--recursive'])
        if run(short+'-clean',['git','diff','--exit-code',sha,'--']): raise RuntimeError('dirty source')
        for file in ['stage3/meter/twice-daily.sh','stage3/meter/entry_overlay.py','stage3/adapt/99-correctness-core-levenshtein/prepare-census.sh']:
            (root/(short+'-'+file.replace('/','_'))).write_bytes((work/file).read_bytes())
        run(short+'-vet',['go','vet','./internal/lower'])
        overlay=pathlib.Path('/tmp/ov-'+short)
        code=run(short+'-overlay',['python3','stage3/census/latent/make_overlay.py',work,overlay])
        if code==0:
            code=run(short+'-rename',['python3',root/'normalize_overlay.py',overlay])
            if code==0:
                data=json.loads((overlay/'overlay.json').read_text())
                (root/(short+'-latent-overlay.json')).write_text(json.dumps(data,indent=2)+'\n')
                files=sorted(set(v for v in data['Replace'].values() if v.endswith('.go.txt')))
                run(short+'-gofmt',['gofmt','-w',*files])
                run(short+'-latent-build',['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o','/tmp/latent','./stage3/census/latent/tool'])
                entry=pathlib.Path('/tmp/entry-ov-'+short)
                code=run(short+'-entry-overlay',['python3','stage3/meter/entry_overlay.py',overlay,entry])
                if code==0:
                    code=run(short+'-entry-rename',['python3',root/'normalize_overlay.py',entry])
                    if code==0:
                        data=json.loads((entry/'overlay.json').read_text())
                        (root/(short+'-entry-overlay.json')).write_text(json.dumps(data,indent=2)+'\n')
                        files=sorted(set(v for v in data['Replace'].values() if v.endswith('.go.txt')))
                        run(short+'-entry-gofmt',['gofmt','-w',*files])
                        run(short+'-entry-build',['go','build','-buildvcs=false','-overlay='+str(entry/'overlay.json'),'-o','/tmp/entry-census-'+short,'./stage3/census/latent/tool'])
finally:
    run('99-restore',['git','switch','--detach',restore])
    run('99-clean',['git','diff','--exit-code',restore,'--'])
print(json.dumps({k:v['exit'] for k,v in results.items()},indent=2))
