import json,pathlib,collections
s=pathlib.Path(__file__).parent

def read(name):
 rows=[json.loads(x) for x in (s/name).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full'
 assert rows[0]['module_format']=='CommonJSBundle'
 assert rows[-1]['status']=='relation_report'
 findings=[f for r in rows for f in (r.get('findings') or [])]
 # Byte spans are observations rather than finding identity: adding the type
 # shifts later source offsets, without changing their locations or findings.
 def key(f):return tuple(f.get(x,'') for x in ('kind','where','reason','text','phase','unit','pass','piece_atom','root_kind','depth','boundary_where','site_where','site_kind'))
 return {key(f):f for f in findings},rows[0]
a,ha=read('main.jsonl');b,hb=read('change.jsonl');m,hm=read('restoring-mutant.jsonl')
gone=[a[k] for k in sorted(a.keys()-b.keys())];new=[b[k] for k in sorted(b.keys()-a.keys())]
# Changed means the same site/phase/unit now has a different reason or kind.
def site(f):return tuple(f.get(x,'') for x in ('where','phase','unit','pass','piece_atom'))
changed=[]
for x in gone[:]:
 peers=[y for y in new if site(y)==site(x)]
 if len(peers)==1:
  y=peers[0];changed.append({'before':x,'after':y});gone.remove(x);new.remove(y)
result={'identity_fields':['kind','where','reason','text','phase','unit','pass','piece_atom','root_kind','depth','boundary_where','site_where','site_kind'],'main_findings':len(a),'change_findings':len(b),'new':new,'gone':gone,'changed':changed,'restoring_mutant_same_findings':a.keys()==m.keys(),'checker_diagnostics':{'main':ha.get('diagnostics'),'change':hb.get('diagnostics'),'mutant':hm.get('diagnostics')}}
(s/'delta.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
