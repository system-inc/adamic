from pathlib import Path
import ast,json,hashlib,os,re,types
D=Path('/workspace/scratch/recordreplay');R=D/'head';O=D/'inventory-audit';catalog=json.loads((O/'catalog.json').read_text());rows=[]
class Retarget(ast.NodeTransformer):
 def visit_Call(self,node):
  text=ast.unparse(node.func.value) if isinstance(node.func,ast.Attribute) else ''
  node=self.generic_visit(node)
  if isinstance(node.func,ast.Attribute) and node.func.attr in ['rglob','glob','iterdir']:
   node.func.value=ast.Name(id='__compiler' if 'src/compiler' in text else '__src',ctx=ast.Load())
  if isinstance(node.func,ast.Attribute) and node.func.attr=='relative_to':
   original=ast.unparse(node.args[0]) if node.args else ''
   if node.args:node.args[0]=ast.Name(id='__compiler' if 'src/compiler' in original else '__src' if 'src' in original else '__tree',ctx=ast.Load())
  return node
for entry in catalog['readers']:
 if not entry['file'].endswith('.py'):continue
 text=(R/entry['file']).read_text()
 try:module=ast.parse(text)
 except SyntaxError:continue
 found=[]
 # Every literal directory comprehension is a directly executable inventory projection.
 for n in ast.walk(module):
  if not isinstance(n,(ast.ListComp,ast.SetComp,ast.DictComp,ast.GeneratorExp)):continue
  if not any(isinstance(x,ast.Call) and isinstance(x.func,ast.Attribute) and x.func.attr in ['rglob','glob','iterdir'] for x in ast.walk(n)):continue
  source=ast.get_source_segment(text,n)
  if not source or not any(k in source for k in ['src','root.rglob','tree.rglob','directory.rglob']):continue
  if any(k in source for k in ['write_','unlink(','rename(','mkdir(','subprocess.','run(']):continue
  row={'file':entry['file'],'line':n.lineno,'method':'original directory comprehension, directory operands retargeted to measured src; unrelated driver not executed','source':source,'results':{}}
  for side,tree in [('control',D/'control'),('amended',D/'candidate2')]:
   ns={'Path':Path,'hashlib':hashlib,'json':json,'os':os,'str':str,'sorted':sorted,'hash':lambda b:hashlib.sha256(b).hexdigest(),'h':lambda b:hashlib.sha256(b).hexdigest(),'sha':lambda b:hashlib.sha256(b).hexdigest(),'digest':lambda p:hashlib.sha256(p.read_bytes() if isinstance(p,Path) else p).hexdigest(),'hashfile':lambda p:hashlib.sha256(p.read_bytes()).hexdigest(),'MEASUREMENT':'measured on a checker-rejected program','__tree':tree,'__src':tree/'src','__compiler':tree/'src/compiler'}
   for k in ['root','tree','adapted','source','before','after','scoped','t','s','D','R','repo','directory','cache','CACHE']:ns[k]=tree
   try:
    expr=Retarget().visit(ast.parse(source,mode='eval'));ast.fix_missing_locations(expr);value=eval(compile(expr,entry['file'],'eval'),ns)
    if isinstance(value,types.GeneratorType):value=list(value)
    names=[]
    def collect(value):
     if isinstance(value,(str,Path)):
      if len(str(value))>1000 or '\n' in str(value):return
      try:
       p=Path(str(value));p=p if p.is_absolute() else tree/p if (tree/p).is_file() else tree/'src'/p
       if p.is_file() and p.is_relative_to(tree/'src'):names.append(str(p.relative_to(tree)))
      except OSError:pass
     elif isinstance(value,dict):
      for k,v in value.items():collect(k);collect(v)
     elif isinstance(value,(list,set,tuple)):
      for v in value:collect(v)
    collect(value)
    row['results'][side]={'status':'pass','source_files':len(set(names)),'contains_added':'src/recordreplay.d.ts' in names,'result_entries':len(value),'inventory_sha256':hashlib.sha256('\n'.join(sorted(set(names))).encode()).hexdigest()}
   except Exception as e:row['results'][side]={'status':'context_required','error':type(e).__name__+': '+str(e)[:240]}
  found.append(row)
 rows+=found
 if not found:rows.append({'file':entry['file'],'method':'no direct src comprehension; retained for caller/exclusion review','results':{}})
(O/'python-readers.json').write_text(json.dumps(rows,indent=2)+'\n');print('Python inventory projections',len(rows),'context required',sum(r.get('results',{}).get('amended',{}).get('status')=='context_required' for r in rows))
