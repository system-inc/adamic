import json,re,hashlib,pathlib,sys
root=pathlib.Path(sys.argv[1]); mode=sys.argv[2]
def receiver(s):
 s=re.sub(r'^([ \t]*)const \{ setTimeout \} = timerHost;\n','',s,flags=re.M)
 return re.sub(r'setTimeout(!?)\.call\(timerHost, ',r'timerHost.setTimeout\1(',s)
def writepins(rel,transform):
 p=root/rel;data=json.loads(p.read_text());count=0
 for pin in data:
  for state in ['before','after']:
   old=pin[state];new=transform(old)
   if old!=new:
    pin[state]=new;pin['owner'+state.title()+'Sha256']=hashlib.sha256(new.encode()).hexdigest();count+=old.count('.call(timerHost,')
 p.write_text(json.dumps(data,indent=2)+'\n');return count
if mode=='1':
 assert writepins('stage3/adapt/99-correctness-timer-closures/owners.json',receiver)==7
 assert writepins('stage3/adapt/99-correctness-axolotl-x/owners.json',receiver)==2
 p=root/'stage3/adapt/99-correctness-axolotl-x/adapt.cjs';s=p.read_text()
 # The two replacements are JSON string literals in this adapter.
 n=0
 def repl(m):
  global n
  raw=m.group();v=json.loads(raw);new=receiver(v)
  if new!=v:n+=1;return json.dumps(new)
  return raw
 s=re.sub(r'"(?:[^"\\]|\\.)*"',repl,s);assert n==2;p.write_text(s)
 p=root/'stage3/adapt/99-correctness-timer-closures/calls.json';a=json.loads(p.read_text())
 assert sum('.call(timerHost,' in x['after'] for x in a)==9
 for x in a:
  if '.call(timerHost,' in x['after']:
   x['after']=receiver(x['after']);x['callForm']='receiver timerHost.setTimeout'+('!(...)' if 'setTimeout!(' in x['after'] else '(...)')
 p.write_text(json.dumps(a,indent=2)+'\n')
else:raise ValueError(mode)
