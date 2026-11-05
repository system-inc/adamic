import collections,gzip,hashlib,json,pathlib,sys
reference,other,output=sys.argv[1:]
def identities(name):
 raw=collections.Counter()
 for line in pathlib.Path(name).read_text().splitlines():
  r=json.loads(line);raw[(r['event'],r['title'])]+=1
 assert all(n%2==0 for n in raw.values()),'uneven host/worker identity multiplicity'
 return raw,collections.Counter({key:n//2 for key,n in raw.items()})
a,normal=identities(reference);b,otherNormal=identities(other)
added=b-a;removed=a-b
result={'rawEventsBefore':sum(a.values()),'rawEventsAfter':sum(b.values()),'normalizedOccurrences':sum(normal.values()),'rawMultisetsMatch':a==b,'normalizedMultisetsMatch':normal==otherNormal,'added':sum(added.values()),'removed':sum(removed.values()),'outcomes':{event:sum(n for (e,t),n in normal.items() if e==event) for event in ['pass','fail','pending']},'failures':[{'title':title,'count':n} for (e,title),n in normal.items() if e=='fail']}
assert a==b,result
assert result['outcomes']=={'pass':106366,'fail':1,'pending':0},result
assert result['normalizedOccurrences']==106367
payload=''.join(json.dumps({'event':e,'title':t,'count':n},ensure_ascii=False,separators=(',',':'))+'\n' for (e,t),n in sorted(normal.items())).encode()
result['normalizedSha256']=hashlib.sha256(payload).hexdigest()
pathlib.Path(output).write_text(json.dumps(result,indent=2)+'\n')
with gzip.open(output+'.identities.jsonl.gz','wb') as f:f.write(payload)
print(json.dumps(result,indent=2))
