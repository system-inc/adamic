from pathlib import Path
import json,re,collections,difflib,subprocess as s,gzip
r=Path(__file__).resolve().parent
# Messages lose source positions only; kind, phase and reason remain exact.
def strip_positions(text):
 return re.sub(r'(?:[A-Za-z]:)?(?:[^\s:;()\[\]"\']*/)?[\w.-]+\.(?:ts|a):\d+:\d+(?:-\d+(?::\d+)?)?:?\s*','',text)
def collect(side):
 rows=[json.loads(l) for l in (r/('census-'+side+'.jsonl')).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle';assert rows[-1]['status']=='relation_report'
 unique={}
 for row in rows:
  for f in row.get('findings',[]):
   identity=tuple(f.get(k,'') for k in ['kind','phase','where','reason','text'])
   unique.setdefault(identity,f)
 return rows,list(unique.values())
def key(f):return tuple(f.get(k,'') for k in ['kind','phase','reason'])+(strip_positions(f.get('text','')),)
base,a=collect('before');after,b=collect('after');left=collections.defaultdict(list);right=collections.defaultdict(list)
for f in a:left[key(f)].append(f)
for f in b:right[key(f)].append(f)
paired=0;gone=[];new=[]
for k in left.keys()|right.keys():
 n=min(len(left[k]),len(right[k]));paired+=n;gone+=left[k][n:];new+=right[k][n:]
# Classify unmatched same-site observations as changed, after mapping unchanged lines.
beforetext=s.check_output(['git','-C',str(r/'before'),'show','refs/stage3/99-wasm-crypto-import-after:src/compiler/sys.ts'],text=True)
aftertext=(r/'after/src/compiler/sys.ts').read_text();mapping={}
for tag,a1,a2,b1,b2 in difflib.SequenceMatcher(a=beforetext.splitlines(),b=aftertext.splitlines(),autojunk=False).get_opcodes():
 if tag=='equal':mapping.update({j+1:a1+j-b1+1 for j in range(b1,b2)})
def location(f,side):
 m=re.search(r':(\d+):(\d+)$',f['where']);assert m
 line=int(m[1]);return (mapping.get(line,('new',line)) if side=='after' else line,int(m[2]),f['kind'],f['phase'])
changed=[]
for old in list(gone):
 match=next((f for f in new if location(old,'before')==location(f,'after')),None)
 if match is not None:gone.remove(old);new.remove(match);changed.append({'before':old,'after':match})
# Locate the original stock sys.ts:1532 watchFile property in the adapted nodeSystem.
stock=gzip.decompress(s.check_output(['git','-C','/workspace/adamic','show','ab3948bc8cb2db29511dcba476209b3d3d1c1739:watch-functions-1532-evidence/cead-check-1532/sys-main.ts.gz'])).decode().splitlines();assert stock[1531].strip()=='watchFile,'
lines=beforetext.splitlines();start=next(i for i,l in enumerate(lines) if 'const nodeSystem:' in l);targetline=next(i+1 for i in range(start,len(lines)) if lines[i].strip()=='watchFile,')
target=[f for f in a if re.search(r':'+str(targetline)+r':13$',f['where']) and f['kind']=='Refused' and f['phase']=='refusal_scan'];assert len(target)==1,target
assert any(f==target[0] for f in gone),'target did not disappear'
result={'pairingKey':['kind','phase','reason','message with source positions stripped'],'beforeUniqueFindings':len(a),'afterUniqueFindings':len(b),'paired':paired,'gone':gone,'new':new,'changed':changed,'target':{'originalSite':'sys.ts:1532:13','reference':'original cead-check-1532/sys-main.ts.gz','adaptedBaselineSite':'sys.ts:'+str(targetline)+':13','sourceProperty':'nodeSystem.watchFile','gone':True},'beforeCheckerDiagnostics':base[0]['diagnostic_sites'],'afterCheckerDiagnostics':after[0]['diagnostic_sites']}
(r/'census-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:len(result[k]) if isinstance(result[k],list) else result[k] for k in ['beforeUniqueFindings','afterUniqueFindings','paired','gone','new','changed','target']},indent=2))
assert not new and not changed,result
