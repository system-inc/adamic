import pathlib,subprocess,os,json,shutil,argparse
parser=argparse.ArgumentParser()
parser.add_argument('--artifacts',required=True);parser.add_argument('--stage0',required=True)
args=parser.parse_args()
source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
def run(name,cmd,expected=0):
 with open(root/(name+'.stdout'),'wb') as out,open(root/(name+'.stderr'),'wb') as err:p=subprocess.run(cmd,stdout=out,stderr=err,cwd=repo)
 assert p.returncode==expected,(name,p.returncode,(root/(name+'.stderr')).read_text())
 return (root/(name+'.stdout')).read_bytes(),(root/(name+'.stderr')).read_bytes()
run('control',['go','test','./stage1/cohere/typeaware/wave08-next','-count=1','-v'])
for name,file,before,after,message in [('return-flags','resolved_callee.go','flags = uint64(returned.Flags())','flags = 0','resolved callee differs'),('module-target','program_loads.go','target = found.FileName()','target = ""','module resolutions absent')]:
 path=source/file;text=path.read_text();assert text.count(before)==1
 replacement=root/(name+'.go');replacement.write_text(text.replace(before,after));overlay=root/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(path):str(replacement)}}))
 output,errors=run(name,['go','test','-overlay',str(overlay),'./stage1/cohere/typeaware/wave08-next','-run','^TestPendingFactsAndRegistrationRefusals$','-count=1','-v'],1)
 assert message.encode() in output and b'build failed' not in output+errors
 print(name,'compiled; direct checker test caught',message,flush=True)
stage=args.stage0
run('flow-build',[stage,'build',str(source/'testdata/flow_probe.a'),'-o',str(root/'flow'),'--sanitize'])
truth,_=run('flow-go',['go','run',str(source/'testdata/flow_oracle.go')]);output,errors=run('flow',[str(root/'flow')]);assert output==truth and errors==b''
print('native event engine agrees with Go expectations under sanitizers',flush=True)
path=source/'symbol_ancestry.go';text=path.read_text();before='name = named.Text()';assert text.count(before)==1
replacement=root/'ancestor-name.go';replacement.write_text(text.replace(before,'name = "mutant"'));overlay=root/'ancestor-name.json';overlay.write_text(json.dumps({'Replace':{str(path):str(replacement)}}))
output,errors=run('ancestor-name',['go','test','-overlay',str(overlay),'./stage1/cohere/typeaware/wave08-next','-run','^TestRawSymbolAncestry$','-count=1','-v'],1)
assert b'ancestor differs' in output and b'build failed' not in output+errors
print('ancestor-name compiled; pinned checker ancestry witnesses catch it',flush=True)

for name,before,after in [('write-state','state.chains.length > 0','state.chains.length > 1'),('blocking-state',"blocking && event.kind === 'block'","false && event.kind === 'block'")]:
 directory=root/name;directory.mkdir(exist_ok=True)
 text=(source/'process_flow.a').read_text();assert text.count(before)==1;(directory/'process_flow.a').write_text(text.replace(before,after))
 (directory/'probe.a').write_text((source/'testdata/flow_probe.a').read_text().replace("'../process_flow.a'","'./process_flow.a'"))
 run(name+'-build',[stage,'build',str(directory/'probe.a'),'-o',str(directory/'mutant')])
 output,errors=run(name,[str(directory/'mutant')]);assert output!=truth and errors==b''
 print(name,'compiled, exit 0, empty stderr; Go event expectations differ',flush=True)
 (directory/'mutant').unlink()
print('PASS',flush=True)
