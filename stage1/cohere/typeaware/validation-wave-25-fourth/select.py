import collections,json,re,subprocess

def git(*args):return subprocess.check_output(['git',*args],text=True)
counts=collections.Counter()
for pop in ['compiler','repository']:
 for line in open('stage1/cohere/typeaware/validation-volume/'+pop+'-all.counts'):
  parts=line.rstrip('\n').split('\t')
  if len(parts)==2 and parts[1].isdigit():counts[parts[0]]+=int(parts[1])
ranking=sorted(counts,key=lambda name:(-counts[name],name));patterns={name:re.compile(r'(?<![\w/@-])'+re.escape(name)+r'(?![\w/-])') for name in ranking}
refs=git('for-each-ref','--format=%(refname)','refs/remotes/origin').splitlines();claims={};ports={};blobs={}
for ref in refs:
 for line in git('ls-tree','-r',ref,'--','stage1/cohere/typeaware/claims').splitlines():
  metadata,path=line.split('\t',1);kind,oid=metadata.split()[1:]
  if kind=='blob' and not path.endswith('.gz'):blobs.setdefault(oid,[]).append({'ref':ref,'path':path})
for oid,locations in blobs.items():
 content=git('show',oid)
 for name,pattern in patterns.items():
  if pattern.search(content):claims.setdefault(name,[]).extend(locations)
for ref in ['refs/remotes/origin/main','refs/remotes/origin/codex/tsgo-c-library']:
 for line in git('ls-tree','-r',ref,'--','stage1/cohere').splitlines():
  metadata,path=line.split('\t',1);kind,oid=metadata.split()[1:]
  if kind!='blob' or not path.endswith(('.a','.ts')) or any(part in path for part in ['/testdata/','/validation','/claims/']):continue
  content=git('show',oid)
  for name,pattern in patterns.items():
   short = name.removeprefix('@typescript-eslint/')
   port_pattern = re.compile(r"(?<![\w/@-])"+re.escape(short)+r"(?![\w/-])") if short != name else pattern
   if port_pattern.search(content):ports.setdefault(name,[]).append({'ref':ref,'path':path})
available=[name for name in ranking if name not in claims and name not in ports]
result={'selected':available[:3],'available':available,'origin_refs':{ref:git('rev-parse',ref).strip() for ref in refs},'distinct_claim_blobs':len(blobs),'ranking':[{'rule':name,'volume':counts[name],'ported':ports.get(name,[]),'claimed':claims.get(name,[])} for name in ranking]}
json.dump(result,open('/workspace/wave-25-fourth-selection.json','w'),indent=2)
print(json.dumps({'refs':len(refs),'claim_blobs':len(blobs),'rules':len(ranking),'ported':len(ports),'claimed':len(claims),'available':available},indent=2))
