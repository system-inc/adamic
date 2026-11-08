import json, os, pathlib, shlex, subprocess, sys
root=pathlib.Path('/workspace/adamic/evidence/l5')
work=pathlib.Path('/workspace/adamic-rule-g-2594029')
head='e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1'
base='e3aa0bb7642163ebd004a71ba40a4657d27eab71'
results={}
def run(label,args,cwd=work,env=None):
    (root/(label+'.command')).write_text(shlex.join(map(str,args))+'\n')
    r=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    (root/(label+'.stdout')).write_bytes(r.stdout)
    (root/(label+'.stderr')).write_bytes(r.stderr)
    code=r.returncode if r.returncode>=0 else 128-r.returncode
    (root/(label+'.exit')).write_text(str(code)+'\n')
    results[label]={'exit':code,'signal':-r.returncode if r.returncode<0 else None,'stdout':r.stdout.decode(errors='replace'),'stderr':r.stderr.decode(errors='replace')}
    (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
    return code,r.stdout
run('00-node-version',['node','--version'])
run('01-go-version',['go','version'])
run('02-clang-version',['clang','--version'])
run('03-source-node',['node','--experimental-strip-types',root/'l5.ts'])
run('04-stock-tsc-version',['node','/workspace/typescript-source/lib/tsc.js','--version'])
run('04-stock-tsc-revision',['git','-C','/workspace/typescript-source','rev-parse','HEAD'])
run('04-stock-tsc-clean',['git','-C','/workspace/typescript-source','diff','--exit-code'])
run('05-stock-tsc-default',['node','/workspace/typescript-source/lib/tsc.js','--ignoreConfig','--noEmit',root/'l5.ts'])
run('06-stock-tsc-strict',['node','/workspace/typescript-source/lib/tsc.js','--ignoreConfig','--noEmit','--strict',root/'l5.ts'])
try:
    for short,sha in [('e5531c8',head),('e3aa0bb',base)]:
        if run(short+'-checkout',['git','switch','--detach',sha])[0]: raise RuntimeError('checkout failed')
        run(short+'-revision',['git','rev-parse','HEAD'])
        run(short+'-submodules',['git','submodule','status','--recursive'])
        if run(short+'-clean',['git','diff','--exit-code',sha,'--'])[0]: raise RuntimeError('dirty source')
        compiler='/tmp/adamic-'+short
        if run(short+'-compiler-build',['go','build','-o',compiler,'./cmd/adamic'])[0]: continue
        run(short+'-help',[compiler,'--help'])
        for ext in ['ts','a']:
            label=short+'-'+ext
            binary='/tmp/l5-'+label
            code,_=run(label+'-native-build',[compiler,'build',root/('l5.'+ext),'-o',binary,'--sanitize'])
            if code==0:
                env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='print_stacktrace=1:halt_on_error=1')
                (root/(label+'-native-run.env')).write_text('ASAN_OPTIONS=detect_leaks=1:halt_on_error=1\nUBSAN_OPTIONS=print_stacktrace=1:halt_on_error=1\n')
                run(label+'-native-run',[binary],env=env)
            code,js=run(label+'-js-build',[compiler,'js',root/('l5.'+ext)])
            if code==0:
                generated=root/(label+'.mjs')
                generated.write_bytes(js)
                run(label+'-js-run',['node',work/'oracle/node.mjs',generated])
finally:
    run('99-restore-head',['git','switch','--detach',head])
    run('99-restoration',['git','diff','--exit-code',head,'--'])
print(json.dumps({k:{'exit':v['exit'],'signal':v['signal']} for k,v in results.items()},indent=2))
