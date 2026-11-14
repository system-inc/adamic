from pathlib import Path
import subprocess,re,json,ast,hashlib
R=Path('/workspace/scratch/recordreplay/head');O=Path('/workspace/scratch/recordreplay/inventory-audit')
paths=subprocess.check_output(['git','ls-files'],cwd=R,text=True).splitlines();entries=[];fixed=[]
for name in paths:
 p=R/name
 if not p.is_file() or not name.endswith(('.py','.cjs','.mjs','.js','.go','.go.txt','.sh')):continue
 if any(x in p.parts for x in ['node_modules','runs']) or '/testdata/' in name:continue
 try:text=p.read_text()
 except (UnicodeError,OSError):continue
 enum=re.search(r'rglob\(|\.glob\(|\.iterdir\(|os\.walk\(|readdirSync\(|(?:filepath\.)?WalkDir\(|os\.ReadDir\(',text)
 if not enum:continue
 relevant=bool(re.search(r"['\"]src(?:/|['\"])|src/compiler|source.inventory|source_manifest|sourceManifest|sourceFiles|payload_digest",text))
 if not relevant:continue
 lines=[{'line':i,'text':l.strip()} for i,l in enumerate(text.splitlines(),1) if re.search(r'rglob\(|\.glob\(|\.iterdir\(|os\.walk\(|readdirSync\(|WalkDir\(|ReadDir\(',l)]
 entries.append({'file':name,'sha256':hashlib.sha256(text.encode()).hexdigest(),'sites':lines})
 if name.endswith('.py'):
  try:tree=ast.parse(text)
  except SyntaxError:continue
  for n in ast.walk(tree):
   if isinstance(n,ast.Assert) and any(isinstance(c,ast.Constant) and c.value==746 for c in ast.walk(n)):
    fixed.append({'file':name,'line':n.lineno,'source':ast.get_source_segment(text,n)})
(O/'catalog.json').write_text(json.dumps({'tracked_files_searched':len(paths),'readers':entries,'fixed_python_count_guards':fixed},indent=2)+'\n')
print('reader candidates',len(entries),'fixed Python guards',fixed)
