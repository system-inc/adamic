import json, pathlib, subprocess
root=pathlib.Path('/workspace/adamic'); ev=root/'evidence'; scratch=pathlib.Path('/tmp/adamic-probes/probe/generic_methods_2')
logs=ev/'probes'; logs.mkdir(exist_ok=True)
records=[]
def run(label,args,stdout=None):
    (logs/(label+'.command.json')).write_text(json.dumps(args)+'\n')
    out=logs/(label+'.stdout'); err=logs/(label+'.stderr')
    with out.open('wb') as o, err.open('wb') as e:
        try:
            rc=subprocess.run(args,stdout=o,stderr=e,timeout=120).returncode
        except subprocess.TimeoutExpired: rc=124
    (logs/(label+'.exit')).write_text(str(rc)+'\n')
    return {'exit':rc,'stdout':out.read_text(errors='replace'),'stderr':err.read_text(errors='replace')}
for src in sorted(scratch.glob('*')):
    if src.suffix not in ('.a','.ts'): continue
    node=run(src.name+'-node',['node',str(src)] if src.suffix=='.ts' else ['node',str(root/'oracle/node.mjs'),str(src)])
    for rev,compiler in [('head','/tmp/adamic'),('base','/tmp/adamic-base')]:
        label=rev+'-'+src.name; binary='/tmp/probe-'+label
        native=run(label+'-build',[compiler,'build',str(src),'-o',binary,'--sanitize'])
        nr=run(label+'-native',[binary]) if native['exit']==0 else None
        js=run(label+'-js',[compiler,'js',str(src)])
        jr=None
        if js['exit']==0:
            target=scratch/(label+'.mjs'); target.write_text(js['stdout'])
            jr=run(label+'-javascript',['node',str(root/'oracle/node.mjs'),str(target)])
        records.append(dict(program=src.name,revision=rev,node=node,native_compile=native,native_run=nr,js_compile=js,js_run=jr))
        (ev/'probe-results.json').write_text(json.dumps(records,indent=2)+'\n')
lines=['# Generic method probes','','Both extensions are tested separately. Native builds use `--sanitize`; JavaScript is emitted with `adamic js` and run through `oracle/node.mjs`. All commands, stdout, stderr, and exits are retained under `evidence/probes/`.','','| Program | Revision | Native compile/run | JS compile/run | Source Node |','|---|---|---|---|---|']
def desc(c,r):
    if c['exit']: return str(c['exit'])+': '+c['stderr'].strip().replace('\n','<br>').replace('|','\\|')
    return '0 / '+str(r['exit'])+': '+repr(r['stdout']).replace('|','\\|')+((' stderr: '+r['stderr'].strip().replace('\n','<br>').replace('|','\\|')) if r['exit'] else '')
for x in records:
    n=x['node']; lines.append('| '+x['program']+' | '+x['revision']+' | '+desc(x['native_compile'],x['native_run'])+' | '+desc(x['js_compile'],x['js_run'])+' | '+str(n['exit'])+': '+repr(n['stdout'])+' |')
(ev/'probes.md').write_text('\n'.join(lines)+'\n')
