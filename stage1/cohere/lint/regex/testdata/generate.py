"""Translate pinned fixed expressions once, retaining all dynamic sites as contracts."""
import json,re,pathlib,collections,hashlib
root=pathlib.Path(__file__).resolve().parents[1]
rows=json.loads((root/'sites.json').read_text())
def translate(p):
 p=re.sub(r'\\x\{([0-9A-Fa-f]+)\}',lambda m:'\\u{'+m[1]+'}',p)
 flags='gu'; insensitive=p.startswith('(?i)'); multi=p.startswith('(?m)'); dotall=p.startswith('(?s)')
 if insensitive:flags+='i';p=p[4:]
 if multi or dotall:p=p[4:]
 out='';inside=False;i=0
 while i<len(p):
  c=p[i]
  if c=='\\':
   i+=1;c=p[i]
   if c=='s':out+=' \\t\\n\\f\\r' if inside else '[ \\t\\n\\f\\r]'
   elif c=='S' and not inside:out+='[^ \\t\\n\\f\\r]'
   elif c=='b' and insensitive and not inside:out+='(?-i:\\b)'
   elif c=='z':out+='(?![\\s\\S])'
   elif c=='-' and not inside:out+='-'
   else:out+='\\'+c
  elif c=='[':inside=True;out+=c
  elif c==']':inside=False;out+=c
  elif c=='.' and not inside:out+='[\\s\\S]' if dotall else '[^\\n]'
  elif c=='$' and not inside:out+='(?=\\n|(?![\\s\\S]))' if multi else '(?![\\s\\S])'
  elif c=='^' and multi and not inside:out+='(?<=\\n|^)'
  else:out+=c
  i+=1
 return out,flags
fixed=[]
for row in rows:
 row['id']=f"{row['file']}:{row['line']}"
 if row['go_pattern'] is not None:
  p,flags=translate(row['go_pattern']);row.update(js_pattern=p,js_flags=flags,status='candidate, differential gate required');fixed.append(row)
 else:row.update(js_constructor="new RegExp(pattern, 'u')", js_pattern=None,js_flags='u',status='dynamic source expression, not a finite pattern; option source uses new RegExp(pattern, u) but Go-only syntax is a gap')
(root/'table.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
lines=['// Generated fixed Go expressions. Translate only at port time.','export function fixedPatterns(): RegExp[] {','    return [']
for r in fixed:
 p=r['js_pattern'].replace('/','\\/').replace('\n','\\n').replace('\r','\\r').replace('\u2028','\\u2028').replace('\u2029','\\u2029')
 lines.append(f"        /{p or '(?:)'}/{r['js_flags']}, // {r['id']}")
lines+=['    ];','}']
for export_name, source_file, source_line in [('inlineCommentDirective','core/no_inline_comments.go',38),('warningSelfDirective','core/no_warning_comments.go',280)]:
 r=next(r for r in fixed if r['file']==source_file and r['line']==source_line)
 pattern=r['js_pattern'].replace('/','\\/')
 lines.append(f"export function {export_name}(): RegExp {{ return /{pattern}/{r['js_flags'].replace('g','')}; }}")
(root/'patterns.a').write_text('\n'.join(lines)+'\n')
# The fleet's named quiet-hundred manifest is not present on main. Keep an explicit,
# reproducible alternate corpus rather than presenting it as that manifest.
compiler=pathlib.Path('/workspace/scratch/wave1-06-typescript-6.0.3/src/compiler')
files=sorted(compiler.rglob('*.ts'))[:77]+sorted((root.parents[3]/'stage1').rglob('*.ts'))[:23]
values=[];meta=[]
for f in files:
 text=f.read_text();meta.append({'path':str(f),'sha256':hashlib.sha256(f.read_bytes()).hexdigest()})
 values+=list(dict.fromkeys(re.findall(r'[A-Za-z_$][A-Za-z0-9_$]*',text)))[:25]
 values+=re.findall(r'//[^\n\r]*|/\*[\s\S]*?\*/',text)[:25]
controls=['PaginationInput','Pagination','','TODO','todo','ſ','K','K','S','ß','é','🌍TODO','TODO🌍','TODO\u00a0x','\u00a0TODO','\vTODO','TODO\n','foo\r','foo\u2028','a\nb','eslint-disable no-warning-comments','var a = 1; // TODO: later','className="flex-grow-2"','inset-shadow','onClick','handleEvent','useEffect(','// XXX']
# Bound deduplicated identifier/comment payloads while retaining every file's contribution.
corpus=list(dict.fromkeys(controls+values))[:5000]
(root/'testdata/corpus.json').write_text(json.dumps(corpus,ensure_ascii=False)+'\n')
(root/'testdata/corpus-files.json').write_text(json.dumps(meta,indent=2)+'\n')
(root/'testdata/corpus.a').write_text('export function inputs(): string[] { return '+json.dumps(corpus,ensure_ascii=False).replace('\u2028','\\u2028').replace('\u2029','\\u2029')+'; }\n')
print('sites',len(rows),'fixed',len(fixed),'dynamic',len(rows)-len(fixed),'files',len(meta),'inputs',len(corpus),'features',dict(collections.Counter(r['feature']for r in rows)))
