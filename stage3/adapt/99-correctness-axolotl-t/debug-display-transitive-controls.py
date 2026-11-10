from pathlib import Path
import ast,copy,json,hashlib
p=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-t/debug-display-classify-hunks.py');source=p.read_text();tree=ast.parse(source);loop=next(n for n in tree.body if isinstance(n,ast.For));body=loop.body
has=lambda n,name:isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id==name for t in n.targets)
start=next(i for i,n in enumerate(body) if has(n,'by_name'));end=next(i for i,n in enumerate(body) if has(n,'drop'));code=compile(ast.fix_missing_locations(ast.Module(body=body[start:end],type_ignores=[])),str(p),'exec')
def declaration(name,line,refs):return dict(name=name,start_line=line,end_line=line+2,original_declaration_line=line,references=refs)
def ref(parent=None,line=50,owner=None):return dict(line=line,column=1,enclosing_function=parent,inside_five_owner=owner)
a=declaration('a',100,[ref(owner='enableDebugInfo')]);b=declaration('b',200,[ref('a',101)]);outside=declaration('a',100,[ref('not_removed',999)]);cycle_a=declaration('a',100,[ref('b',201)]);cycle_b=declaration('b',200,[ref('a',101)]);mixed=declaration('a',100,[ref(owner='enableDebugInfo'),ref('not_removed',999)])
cases=[('direct owner',[a],['a']),('reverse ordered transitive chain',[b,a],['a','b']),('external caller',[outside],[]),('unanchored cycle',[cycle_a,cycle_b],[]),('mixed owner and external caller',[mixed],[]),('chain ending at external caller',[b,outside],[])]
rows=[]
for name,removed,expected in cases:
 scope={'removed':copy.deepcopy(removed)};exec(code,scope);accepted=sorted(scope['accepted']);assert accepted==expected,(name,accepted,expected);rows.append({'case':name,'accepted':accepted,'pass':True})
Path('/workspace/scratch/debug-display-transitive-controls.json').write_text(json.dumps({'status':'pass','actual_classifier_block_executed':True,'classifier_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'rows':rows},indent=2)+'\n');print(json.dumps({'status':'pass','cases':len(rows)}))
