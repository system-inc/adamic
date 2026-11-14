import pathlib,json,re,difflib,sys
p=pathlib.Path(__file__).parent
oldtag,newtag,file,oldcensus,newcensus=sys.argv[1:]
base=(p/f'sources.{oldtag}'/file).read_text().splitlines();after=(p/f'sources.{newtag}'/file).read_text().splitlines()
mapping={j+1:i+1 for tag,i1,i2,j1,j2 in difflib.SequenceMatcher(None,base,after,autojunk=False).get_opcodes()if tag=='equal'for i,j in zip(range(i1,i2),range(j1,j2))}
def findings(name):
 rows=[json.loads(x)for x in (p/name).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle' and rows[-1]['status']=='relation_report'
 assert any(r.get('file','').endswith('/src/'+file) for r in rows),'unloaded selected file'
 return rows[0], [f for r in rows for f in r.get('findings',[])]
def normalize(f,shift):
 def sub(m):
  line=int(m[1]);return file+':'+str(mapping.get(line,'edited'+str(line))if shift else line)+':'+m[2]
 return tuple(re.sub(r'[^\s]*?/src/'+re.escape(file)+r':(\d+):(\d+)',sub,f.get(k,''))for k in ['kind','where','reason','text'])
h0,all0=findings(oldcensus);h1,all1=findings(newcensus)
f0=[f for f in all0 if '/src/'+file+':' in f.get('where','')];f1=[f for f in all1 if '/src/'+file+':' in f.get('where','')]
dep0=[f for f in all0 if '/src/'+file+':' not in f.get('where','')];dep1=[f for f in all1 if '/src/'+file+':' not in f.get('where','')]
def depkey(f):return tuple(re.sub(r'[^\s]*?/src/', 'src/',f.get(k,''))for k in ['kind','where','reason','text'])
d0={depkey(f):f for f in dep0};d1={depkey(f):f for f in dep1}
deps={'old_unique':len(d0),'new_unique':len(d1),'gone':[v for k,v in d0.items()if k not in d1],'new':[v for k,v in d1.items()if k not in d0]}

a={normalize(f,False):f for f in f0};b={normalize(f,True):f for f in f1}
result={'file':file,'old':oldtag,'new':newtag,'original_unique':len(a),'new_unique':len(b),'unchanged_after_line_mapping':len(a.keys()&b.keys()),'gone':[v for k,v in a.items()if k not in b],'new_findings':[v for k,v in b.items()if k not in a],'old_checker_rejected':h0['checker_rejected'],'new_checker_rejected':h1['checker_rejected'],'dependency_findings':deps,'note':'All distinct kind/where/reason/text findings, Boundary included. Changed findings appear in both gone and new. Relocated unchanged lines map to old coordinates. No other file is discarded from raw streams.'}
name=f'delta.{oldtag}-{newtag}.'+pathlib.Path(file).stem+'.json';(p/name).write_text(json.dumps(result,indent=2)+'\n')
print(name,len(a),len(b),'gone',len(result['gone']),'new',len(result['new_findings']))
