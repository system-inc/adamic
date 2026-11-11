import pathlib,json,sys,subprocess,difflib,gzip
sys.dont_write_bytecode=True
follow=pathlib.Path(__file__).resolve().parent;repo=follow.parents[3]
sys.path.insert(0,str(repo/'stage3/lane'))
import check
label=sys.argv[1];target=follow/label;root=pathlib.Path('/workspace/scratch/baobab-lane-api-'+label);tree=root/'adapted-tree';tree.mkdir(parents=True,exist_ok=True)
commands=[]
def run(args):
 p=subprocess.run(args,capture_output=True);commands.append({'command':args,'child_exit':p.returncode,'stderr':p.stderr.decode()});assert p.returncode==0,(args,p.stderr);return p.stdout
pin='050880ce59e30b356b686bd3144efe24f875ebc8'
run(['git','init','-q',str(tree)])
objects=run(['git','-C','/workspace/scratch/baobab-typescript','rev-parse','--path-format=absolute','--git-path','objects']).decode().strip()
(tree/'.git/objects/info/alternates').write_text(objects+'\n');run(['git','-C',str(tree),'update-ref','HEAD',pin])
original=run(['git','-C',str(tree),'show',pin+':tests/baselines/reference/api/typescript.d.ts'])
p=target/'typescript.d.ts';local=p.read_bytes() if p.exists() else gzip.decompress(pathlib.Path(str(p)+'.gz').read_bytes())
for folder,text in [('reference',original),('local',local)]:
 p=tree/'tests/baselines'/folder/'api/typescript.d.ts';p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(text)
(root/'oracle').mkdir(exist_ok=True)
(root/'oracle/baseline.diff').write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),local.decode().splitlines(True),fromfile='reference/api/typescript.d.ts',tofile='local/api/typescript.d.ts')))
errors,facts=check.api_check(root,json.loads((repo/'stage3/lane/sanctioned-api.json').read_text()))
result={'label':label,'actual_lane_api_check':True,'errors':errors,'facts':facts,'setup_commands':commands,'source_pin':pin}
(target/'lane-api-check.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'label':label,'lane_errors':len(errors),'first_errors':errors[:5]}))
