from pathlib import Path
import os,sys,shutil,json,collections
sys.path.insert(0,'/workspace/scratch/sys-walls');from run import run,S,ROOT
before=S/'before';after=S/'candidate';mutant=S/'mutant';mutant.mkdir();shutil.copytree(after/'src',mutant/'src');(mutant/'src/compiler/sys.ts').write_bytes((before/'src/compiler/sys.ts').read_bytes())
for side in ['before','candidate','mutant']:
 tree=S/side;env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')};env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(S/('latent-'+side+'-times.jsonl')))
 code=run('latent-'+side,[S/'bin/latent',tree/'src/compiler',S/('latent-'+side+'.jsonl')],env=env)
 if code:raise SystemExit(code)
beforetext=(before/'src/compiler/sys.ts').read_text();aftertext=(after/'src/compiler/sys.ts').read_text();lines=beforetext.splitlines();deleted=[i+1 for i,line in enumerate(lines) if line.startswith('declare function setTimeout(') or line.startswith('declare function clearTimeout(')];assert len(deleted)==2
assert [line for i,line in enumerate(lines) if i+1 not in deleted]==aftertext.splitlines()
def read(side):
 rows=[json.loads(l) for l in (S/('latent-'+side+'.jsonl')).read_text().splitlines()];assert rows[0]['diagnostics']==[],rows[0];files=[r for r in rows[1:] if str(r.get('file','')).endswith('/sys.ts')];assert len(files)==1,(side,len(files));return rows,files[0]
def key(f,side):
 where=f['where'];parts=where.rsplit(':',2)
 if len(parts)!=3:return (where,f['kind'],f['reason'],f['phase'])
 name,line,col=parts;line=int(line)
 if name.endswith('/sys.ts'):
  if side!='candidate':
   if line in deleted:return ('deleted-declaration',line,col,f['kind'],f['reason'],f['phase'])
   line-=sum(d<line for d in deleted)
  name='sys.ts'
 else:name=name.replace(str(S/side),'<tree>')
 return (name,line,col,f['kind'],f['reason'],f['phase'])
sets={};counts={}
for side in ['before','candidate','mutant']:
 rows,file=read(side);fs=file.get('findings',[]);sets[side]={key(f,side) for f in fs};counts[side]={'checkerDiagnostics':len(rows[0]['diagnostics']),'units':len(file.get('units',[])),'statuses':dict(collections.Counter(u['status'] for u in file.get('units',[]))),'observations':len(fs),'distinctSites':len(sets[side])}
removed=sets['before']-sets['candidate'];added=sets['candidate']-sets['before'];same=sets['mutant']==sets['before']
result={'counts':counts,'deletedDeclarationLinesBefore':deleted,'removed':[list(k) for k in sorted(removed,key=str)],'added':[list(k) for k in sorted(added,key=str)],'mutantRestoresBaseline':same,'ambientDeclarationFindingsBefore':sum(k[0]=='deleted-declaration' and k[-2]=='an ambient function declaration' for k in sets['before']),'ambientDeclarationFindingsAfter':sum(k[-2]=='an ambient function declaration' for k in sets['candidate'])}
(S/'latent-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
if added or not same or not result['ambientDeclarationFindingsBefore'] or result['ambientDeclarationFindingsAfter']:raise SystemExit('Timer census gate failed; no adapter construction')
