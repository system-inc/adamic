import hashlib,json,pathlib,subprocess
base='5e33a17b186a8a2218d27b69b21e2de5acc5b750'
paths=['internal/load','internal/lower','internal/ir','internal/javascript','cohere','go.mod','go.sum']
rows=[]
for path in paths:
 untracked=subprocess.check_output(['git','ls-files','--others','--exclude-standard','--',path],text=True).splitlines()
 assert not untracked,(path,'untracked inputs',untracked)
 p=subprocess.run(['git','diff','--exit-code',base,'--',path],stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 assert p.returncode==0,(path,p.stdout.decode(),p.stderr.decode())
 identity=subprocess.check_output(['git','ls-tree',base,path],text=True).strip()
 rows.append({'path':path,'base_tree_entry':identity,'worktree_equals_base':True})
out={'base':base,'proof':'All parser/checker/load/lower/IR inputs and Go dependencies are identical to main. Native emission only runs after admission. Thus no frontend/lowering admission or refusal can change for any input. Differential runtime behavior is checked separately. This does not claim all native C compile outcomes are unchanged: portability repairs those outcomes.','inputs':rows}
pathlib.Path('review/compiler/chain-slice-1/admission-delta.json').write_text(json.dumps(out,indent=2)+'\n')
print('PASS: exact frontend/lowering input equality; admission delta is empty')
