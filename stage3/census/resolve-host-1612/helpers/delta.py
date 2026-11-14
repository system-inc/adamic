import json,re,difflib,collections
from pathlib import Path
s=Path(__file__).parent;e=s/'evidence';base=(s/'base-sys.ts').read_bytes().decode().splitlines();change=(s/'accepted-change-sys.ts').read_bytes().decode().splitlines()
line_map={}
for block in difflib.SequenceMatcher(a=base,b=change,autojunk=False).get_matching_blocks():
 for i in range(block.size):line_map[block.b+i+1]=block.a+i+1
# The sole edited pre-existing declaration line changes only its type name.
old_decl=[i+1 for i,x in enumerate(base) if 'const nodeSystem: System = {' in x]
new_decl=[i+1 for i,x in enumerate(change) if 'const nodeSystem: NodeResolutionSystem = {' in x]
assert len(old_decl)==len(new_decl)==1
line_map[new_decl[0]]=old_decl[0]
fields=['kind','where','reason','text','phase','unit','pass','piece_atom','root_kind','depth','boundary_where','site_where','site_kind']
def read(name,tree,candidate=False):
 rows=[json.loads(x) for x in (e/name).read_text().splitlines()];assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle';assert rows[-1]['status']=='relation_report'
 def norm(value):
  if not isinstance(value,str):return value
  value=value.replace(str(tree),'TREE')
  if candidate:
   def loc(m):
    line=int(m[1]);assert line in line_map,('unmapped finding location',line,value)
    return 'TREE/src/compiler/sys.ts:'+str(line_map[line])+':'+m[2]
   value=re.sub(r'TREE/src/compiler/sys\.ts:(\d+):(\d+)',loc,value)
  return value
 findings=[f for r in rows for f in (r.get('findings')or [])]
 normalized={tuple(norm(f.get(k,''))for k in fields):f for f in findings}
 return normalized,rows[0]
a,ha=read('census-base.jsonl',s/'base-tree');b,hb=read('census-fields.jsonl',s/'build-change',True);m,hm=read('census-mutant-accepted.jsonl',s/'build-mutant')
gone=[a[k]for k in sorted(a.keys()-b.keys())];new=[b[k]for k in sorted(b.keys()-a.keys())]
assert len(a)==184 and len(b)==183;assert len(gone)==1 and not new;assert ':1612:77' in gone[0]['where'];assert a.keys()==m.keys()
result={'identityFields':fields,'locationNormalization':'exact equal source lines mapped with difflib; only insertion/changed-site lines excluded; byte spans are observations, not identity','baselineFindings':len(a),'candidateFindings':len(b),'gone':gone,'new':new,'changed':[],'restoringMutantFindings':len(m),'restoringMutantExactFindings':True,'checkerDiagnostics':{'baseline':ha['diagnostics'],'candidate':hb['diagnostics'],'mutant':hm['diagnostics']},'sourceLineMap':line_map}
(e/'delta.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items()if k!='sourceLineMap'},indent=2))
