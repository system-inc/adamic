import json,os,pathlib,subprocess,time
out=pathlib.Path(__file__).parent
records=[]
def run(name,args,env=None):
    if any(r['name']==name and r['exit']==0 for r in records):
        print(name,'completed observation retained',flush=True);return
    started=time.monotonic()
    with (out/(name+'.log')).open('w') as log:
        child=subprocess.run(args,env=env,stdout=log,stderr=subprocess.STDOUT)
    row=dict(name=name,argv=args,exit=child.returncode,seconds=time.monotonic()-started)
    records.append(row);(out/'final-root-commands.json').write_text(json.dumps(records,indent=2)+'\n')
    print(name,'child exit='+str(child.returncode),flush=True)
    if child.returncode: raise SystemExit(child.returncode)
for label in ['after-final']:
    tree=pathlib.Path('/workspace/scratch/tracing-'+label)
    roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'))
    assert len(roots)==79
    run(label+'-types',['/workspace/scratch/tracing-adamic','types',*roots])
    config=out/(label+'-census-config.json')
    config.write_text(json.dumps(dict(extends=str(tree/'src/compiler/tsconfig.json'),compilerOptions=dict(strict=True,strictBindCallApply=True,noUncheckedIndexedAccess=True,exactOptionalPropertyTypes=True,verbatimModuleSyntax=True,erasableSyntaxOnly=False,module='ESNext',moduleResolution='Bundler',moduleDetection='force',target='ES2024',lib=['es2024'],noEmit=True,emitDeclarationOnly=False),include=[],files=roots),indent=2)+'\n')
    env=dict(os.environ,CENSUS_CONFIG=str(config),LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(tree/'src/compiler/tracing.ts'))
    run(label+'-full-census',['/workspace/scratch/tracing-latent',str(tree/'src/compiler'),str(out/(label+'-full-census.jsonl'))],env)
    run(label+'-core-first-stop',['/workspace/scratch/tracing-first-stop',str(tree/'src/compiler/core.ts'),str(out/(label+'-core-first-stop.jsonl'))])
