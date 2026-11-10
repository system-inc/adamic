from pathlib import Path
import json,hashlib,subprocess,os
s=Path('/workspace/scratch');candidate=s/'debug-display-candidate';applied=s/'debug-display-final-tree';view=s/'debug-display-final-lane-view';u=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-t')
def inventory(root):return {str(p.relative_to(root)):p for p in sorted((root/'src').rglob('*')) if p.is_file()}
a=inventory(candidate);b=inventory(applied);assert a.keys()==b.keys(),(len(a),len(b))
source_hashes=[]
for name in a:
 left=a[name].read_bytes();right=b[name].read_bytes();assert left==right,name;source_hashes.append([name,hashlib.sha256(right).hexdigest()])
comparison=json.loads((u/'debug-display-normal-comparison.json').read_text());js=[]
for row in comparison['javascript']:
 p=candidate/'built/local'/row['file'];digest=hashlib.sha256(p.read_bytes()).hexdigest();assert digest==row.get('after_sha256',row.get('sha256'));assert p.read_bytes()==(view/'built/local'/row['file']).read_bytes();js.append({'file':row['file'],'sha256':digest,'matches_classified_candidate_and_lane':True})
graph_code="const fs=require('fs'),crypto=require('crypto'),assert=require('node:assert/strict'),{owners}=require('./stage3/adapt/99-correctness-axolotl-t/adapt.cjs');const a=owners(fs.readFileSync(process.argv[1],'utf8'),'debug.ts').get('Debug.formatControlFlowGraph').text,b=owners(fs.readFileSync(process.argv[2],'utf8'),'debug.ts').get('Debug.formatControlFlowGraph').text;assert.equal(a,b);console.log(crypto.createHash('sha256').update(a).digest('hex'));"
graph=subprocess.run(['node','-e',graph_code,str(s/'debug-display-before/src/compiler/debug.ts'),str(applied/'src/compiler/debug.ts')],cwd='/workspace/adamic',env=dict(os.environ,NODE_PATH='/workspace/adamic/stage3/api/node_modules'),capture_output=True,text=True);print('FlowGraphNode owner invariance child exit:',graph.returncode);assert graph.returncode==0,graph.stderr
report={'FlowGraphNode_owner_sha256':graph.stdout.strip(),'status':'pass','apply_child_exit':0,'apply_command':'bash stage3/apply.sh /workspace/scratch/debug-display-final-tree','source_files':len(a),'complete_applied_source_byte_identical_to_normal_candidate':True,'source_inventory_sha256':hashlib.sha256(json.dumps(source_hashes,separators=(',',':')).encode()).hexdigest(),'normal_javascript':js,'apply_log_sha256':hashlib.sha256((s/'debug-display-final-apply.log').read_bytes()).hexdigest(),'adapter_sha256':hashlib.sha256((u/'adapt.cjs').read_bytes()).hexdigest(),'pins_sha256':hashlib.sha256((u/'pins.json').read_bytes()).hexdigest(),'main':subprocess.check_output(['git','rev-parse','origin/main'],cwd='/workspace/adamic',text=True).strip(),'FlowGraphNode_unchanged':True}
(s/'debug-display-final-source-verification.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps({k:v for k,v in report.items() if k!='normal_javascript'}))
