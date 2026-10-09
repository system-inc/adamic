#!/usr/bin/env python3
"""Prove closed-replay guards and the scout's independent process comparisons."""
import argparse,json,os,subprocess,sys
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('scout',type=Path);parser.add_argument('record',type=Path);parser.add_argument('tree',type=Path);parser.add_argument('slice',type=Path);parser.add_argument('cache',type=Path);parser.add_argument('output',type=Path);args=parser.parse_args()
args.output.mkdir();root=Path(__file__).resolve().parent
original=json.loads(args.record.read_text());request=args.cache/'request.json';golden=(args.cache/'golden.stdout').read_bytes()
rows=[]
for name in ['baseline','order','argument','complete','request','source-input','sources','bom','emitSkipped']:
    data=json.loads(json.dumps(original));expected=None
    if name=='order':
        index=next(i for i,c in enumerate(data['calls']) if c['group']=='resolver');data['calls'].pop(index);expected='resolver order mismatch'
    elif name=='argument':
        call=next(c for c in data['calls'] if c['group']=='resolver');call['args'][0]['pos']=0;expected='resolver argument mismatch'
    elif name=='complete':
        data['calls'].append(dict(group='resolver',name='hasGlobalName',args=['unreached'],result=False));expected='resolver transcript incomplete'
    elif name=='source-input':data['context']['sources'][0]['text']=data['context']['sources'][0]['text'].replace('41 + 1','41 + 2')
    record=args.output/(name+'.json');record.write_text(json.dumps(data))
    entry=args.output/(name+'.a')
    with (args.output/(name+'-generate.log')).open('wb') as log:
        subprocess.run(['node',str(root/'make-driver.cjs'),str(record),str(args.tree),str(args.slice),str(entry)],stdout=log,stderr=log,check=True)
    if name in ['sources','bom','emitSkipped']:
        before,after={'sources':('(arg4 ?? [])','[]'),'bom':('bom: arg2','bom: !arg2'),'emitSkipped':('emitSkipped: result.emitSkipped','emitSkipped: !result.emitSkipped')}[name]
        text=entry.read_text();assert before in text;entry.write_text(text.replace(before,after))
    input_path=request
    if name=='request':
        input_path=args.output/'changed-request.json';input_path.write_bytes(request.read_bytes().replace(b'41 + 1',b'41 + 2'));expected='this closed replay requires its recorded input bytes'
    comparison=args.output/(name+'-comparison')
    with (args.output/(name+'.log')).open('wb') as log:
        result=subprocess.run([sys.executable,str(args.scout/'compare.py'),'--output',str(comparison),str(args.cache/'golden.stdout'),'--','node',str(root/'node.mjs'),str(entry),str(input_path)],stdout=log,stderr=log)
    report=json.loads((comparison/'report.json').read_text())
    if name=='baseline':assert result.returncode==0 and report['success']
    elif name in ['source-input','sources','bom','emitSkipped']:assert result.returncode==1 and set(report['differences'])=={'stdout'},report
    else:
        assert result.returncode==1 and expected in (comparison/'actual.stderr').read_text(),(name,report)
    rows.append(dict(name=name,catcher=expected or ('stock emitted stdout bytes' if name=='source-input' else 'baseline'),report=report,comparison_exit=result.returncode))
    print('PASS '+name,flush=True)
# These are explicitly process test doubles, never native evidence.
for name,stdout,stderr,exit_code,failures in [
    ('stdout',golden.replace(b'41 + 1',b'41 + 2',1),b'',0,['stdout']),
    ('stderr',golden,b'x',0,['stderr']),
    ('exit',golden,b'',1,['exit']),
]:
    fixture=args.output/(name+'.stdout');fixture.write_bytes(stdout)
    comparison=args.output/(name+'-comparison')
    script='import pathlib,sys;sys.stdout.buffer.write(pathlib.Path(sys.argv[1]).read_bytes());sys.stderr.write(sys.argv[2]);sys.exit(int(sys.argv[3]))'
    with (args.output/(name+'.log')).open('wb') as log:
        result=subprocess.run([sys.executable,str(args.scout/'compare.py'),'--output',str(comparison),str(args.cache/'golden.stdout'),'--',sys.executable,'-c',script,str(fixture),stderr.decode(),str(exit_code)],stdout=log,stderr=log)
    report=json.loads((comparison/'report.json').read_text());assert result.returncode==1 and sorted(report['differences'])==failures,(name,report)
    rows.append(dict(name='process-'+name,catcher=failures,test_double=True,report=report,comparison_exit=1))
    print('PASS process-'+name,flush=True)
(args.output/'report.json').write_text(json.dumps(rows,indent=2)+'\n')
