import json,itertools,pathlib,sys,gzip
root=pathlib.Path(__file__).resolve().parents[5]
here=pathlib.Path(__file__).resolve().parent
inv=json.loads((here.parents[1]/'inventory/inventory.json').read_text())
rules=[r for r in inv['rules'] if r['wave']=='syntax waiting on helpers' and not any('ported with' in b['status'] or 'partial port:' in b['status'] for b in r['stage1'])]
schemas=json.loads((root/'cohere/internal/lint/optionschema/schemas.json').read_text())['rules']
unsupported={'pattern','patternProperties','contains','dependencies','exclusiveMaximum','exclusiveMinimum','format','if','then','else','maxProperties','minProperties','multipleOf','propertyNames','$data'}
def features(s):
 if not isinstance(s,dict):return set()
 out=set(s)
 for k in ['properties','$defs','definitions']:
  for c in s.get(k,{}).values():out|=features(c)
 for k in ['allOf','anyOf','oneOf']:
  for c in s.get(k,[]):out|=features(c)
 for k in ['not','items','additionalItems','additionalProperties']:
  c=s.get(k)
  if isinstance(c,list):
   for z in c:out|=features(z)
  elif isinstance(c,dict):out|=features(c)
 return out

def seeds(s,doc,depth=0):
 if depth>15:return []
 if not isinstance(s,dict):return [None,True,False,0,'',[],{}]
 if '$ref' in s:
  t=doc
  for p in s['$ref'][2:].split('/'):t=t[int(p)] if isinstance(t,list) else t[p.replace('~1','/').replace('~0','~')]
  return seeds(t,doc,depth+1)
 out=[None,True,False,0,1,-1,1.5,'','x',[],{}]
 out+=s.get('enum',[])
 if 'const' in s:out.append(s['const'])
 for k in ['oneOf','anyOf','allOf']:
  for b in s.get(k,[]):out+=seeds(b,doc,depth+1)
 props=s.get('properties',{})
 if props:
  full={}
  for k,v in props.items():
   vs=seeds(v,doc,depth+1); full[k]=vs[-1] if vs else None
   out += [{k:x} for x in vs]
   out += [{k.upper():True},{k+'Typo':True}]
  out += [full,dict(full,unexpected=True)]
 items=s.get('items')
 if isinstance(items,list):
  parts=[seeds(i,doc,depth+1) for i in items]
  out += [[x] for x in (parts[0] if parts else [])]
  if len(parts)>1:out += [[a,b] for a in parts[0] for b in parts[1]]
  out += [[None]* (len(items)+1)]
 elif isinstance(items,dict):
  for v in seeds(items,doc,depth+1):out += [[v],[v,v]]
 return out[:300]
cases=[]; supported=[]; missing=[]
for r in rules:
 if 'strict option decoding and schema validation' not in r['waiting_on']:continue
 name=r['name'];s=schemas.get(name)
 if s is None or s is False or features(s)&unsupported or name=='boundaries/dependencies':missing.append(name);continue
 supported.append(name)
 seen=set()
 for value in seeds(s,s):
  text=json.dumps(value,ensure_ascii=True,separators=(',',':'))
  if len(text) < 5000 and text not in seen:cases.append(dict(Kind='schema',Rule=name,Schema=json.dumps(s,separators=(',',':')),Input=text));seen.add(text)
samples=json.loads((root/'cohere/internal/lint/optionschema/testdata/samples.json').read_text())['samples']
for sample in samples:
 name=sample['rule']
 if name in supported:cases.append(dict(Kind='schema',Rule=name,Schema=json.dumps(schemas[name],separators=(',',':')),Input=json.dumps(sample['elements'],separators=(',',':')),Expected='valid' if sample['eslint']=='accepts' else 'invalid'))
# Grammar controls plus duplicate keys, escaping, supplementary characters, null and deep equality.
for text in ['null','true','false','0','-0','1e3','1e400','[1,2]','{"x":1,"x":2}','"\\uD800"','"\\uD83D\\uDE00"','"é"','"\\n"','"\\x20"','"\\uZZZZ"','[1,]','{"x":1,}','{x:1}','01','+1','.1','1.','1e','--1','undefined','NaN','/* hi */1','1 2','\u00a01','"\n"','']:
 cases.append(dict(Kind='json',Input=text))
for schema,values in [({'type':'array','uniqueItems':True},['[{"a":1,"b":2},{"b":2,"a":1}]','[1,1.0]','["a","\\u0061"]']),({'type':'array','items':{'const':'😀'}},['["\\uD83D\\uDE00"]','["x"]']),({'type':'array','items':{'type':'string','minLength':2}},['["😀"]','["😀a"]']),({'type':'array','items':{'oneOf':[{'type':'number'},{'minimum':0}]}},['[1]','[-1]']),({'type':'array','items':{'const':'�'}},['["\\uD800"]','["\\uDC00"]']),({'type':'array','uniqueItems':True},['[1e400,1e400]'])]:
 for value in values:cases.append(dict(Kind='schema',Schema=json.dumps(schema),Input=value))
descriptors=json.loads((here/'descriptors.json').read_text())
# A shape a custom decoder states by hand, or a recursive one, is reached below a field's top level, so
# its samples come from the shape itself: each part's accepted and refused values, one part at a time.
handStated=['"anyOf"','"tuple"','"ref"','"enum"','"whole"','"required"','"nonNull"']
def typical(s,defs,depth=0):
 if depth>3:return None
 kind=s['kind']
 if kind=='ref':return typical(defs[s['name']],defs,depth+1)
 if kind=='anyOf':return typical(s['options'][-1],defs,depth+1)
 if kind=='tuple':return [typical(i,defs,depth+1) for i in s['items']]
 if kind=='string':return (s.get('enum') or ['x'])[0]
 if kind=='number':return 1
 if kind=='boolean':return True
 if kind=='array':return [typical(s['item'],defs,depth+1)]
 if 'fields' in s:return {k:typical(f['shape'],defs,depth+1) for k,f in s['fields'].items()}
 return {}
def examples(s,defs,depth=0):
 if depth>6:return [None]
 kind=s['kind']
 if kind=='ref':return examples(defs[s['name']],defs,depth+1)
 out=[None,True,0,'x',[],{}]
 if kind=='anyOf':
  for option in s['options']:out+=examples(option,defs,depth+1)
 elif kind=='tuple':
  full=typical(s,defs)
  out+=[full,full[:-1],full+[None]]
  for i,item in enumerate(s['items']):out+=[full[:i]+[v]+full[i+1:] for v in examples(item,defs,depth+1)]
 elif kind=='string':out+=s.get('enum',[])
 elif kind=='number':
  out+=[1,-1,1.5,2.0]
  if s.get('whole'):m=int(s['magnitude']);out+=[m,-m,m+1,-m-1]
 elif kind=='array':out+=[[v] for v in examples(s['item'],defs,depth+1)]
 elif 'fields' in s:
  full=typical(s,defs)
  out+=[full,dict(full,unexpected=True)]
  for k,f in s['fields'].items():
   rest={j:v for j,v in full.items() if j!=k}
   out+=[rest,dict(rest,**{k.upper():full[k]})]+[dict(rest,**{k:v}) for v in examples(f['shape'],defs,depth+1)]
 elif 'item' in s:out+=[{'k':v} for v in examples(s['item'],defs,depth+1)]
 return out
decodeRules=[]
for r in rules:
 if 'strict option decoding and schema validation' not in r['waiting_on']:continue
 name=r['name'];d=descriptors.get(name)
 if d is None or 'unsupported' in json.dumps(d):continue
 decodeRules.append(name)
 samples=[None,{},[],True,0,'x']
 for key,field in d.get('fields',{}).items():
  for value in [None,True,False,0,1,1.0,'x',[],['x'],{},[{}]]:
   samples.extend([{key:value},{key.upper():value},{key+'Typo':value}])
  if any(word in json.dumps(field) for word in handStated):
   seen=set()
   for value in examples(field['shape'],d.get('definitions',{})):
    text=json.dumps(value,separators=(',',':'))
    if text not in seen:seen.add(text);samples.append({key:value})
 for value in samples:cases.append(dict(Kind='decode',Rule=name,Schema=json.dumps(d,separators=(',',':')),Input=json.dumps(value,separators=(',',':'))))
for name in decodeRules:
 d=descriptors[name]
 for key,field in d.get('fields',{}).items():
  if field['shape'].get('integer'):
   for raw in ['9223372036854775807','9223372036854775808','-9223372036854775808','-9223372036854775809','9007199254740993','1.0','1e0','-0']:
    cases.append(dict(Kind='decode',Rule=name,Schema=json.dumps(d,separators=(',',':')),Input='{'+json.dumps(key)+':'+raw+'}'))
cat=json.loads((here/'catalog.json').read_text())
messageRules=[]
for r in rules:
 if not any('/policy.' in b for b in r['waiting_on']):continue
 name=r['name'];messageRules.append(name)
 for id,t in cat.get(name,{}).items():
  variants=list(itertools.product(*[[(p,n) for n in opts] for p,opts in (t['Options'] or {}).items()]))
  for variant in variants:
   text=t['Text']; choices=[]
   for p,n in variant:
    text=text.replace('<<'+p+'>>',t['Options'][p][n]);choices.append(dict(Rule=name,Id=id,Phrase=p,Name=n))
   import re
   names=sorted(set(re.findall(r'{{([a-zA-Z0-9]+)}}',text)))
   for val in ['sentinel é😀','{{z}} <<reason>> "quote"\n']:
    cases.append(dict(Kind='message',Rule=name,Id=id,Values={n:val for n in names},Choices=choices))
definitions={}
ids={}
for case in cases:
 if 'Schema' in case:
  schema=case['Schema']
  if schema not in ids:
   name='S'+str(len(ids));ids[schema]=name;definitions[name]=schema
  case['Schema']=ids[schema]
corpus={'Definitions':definitions,'Cases':cases}
(here/'cases.json.gz').write_bytes(gzip.compress((json.dumps(corpus,ensure_ascii=True,separators=(',',':'))+'\n').encode(),mtime=0))
(here/'coverage.json').write_text(json.dumps({'supported_schemas':supported,'unsupported_schemas':missing,'message_rules':messageRules,'strict_decode_targets':decodeRules,'cases':len(cases)},indent=2)+'\n')
print(len(cases),'cases;',len(supported),'option schemas;',len(messageRules),'message consumers')
