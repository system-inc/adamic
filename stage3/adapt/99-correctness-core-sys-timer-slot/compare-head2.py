from run import *
import collections,hashlib
labels=['main','head1-candidate','head1-mutant','head2-candidate','head2-mutant'];sets={};summaries={}
for label in labels:
 rows=[json.loads(l) for l in (S/(label+'-latent.jsonl')).read_text().splitlines()];assert rows[0]['diagnostics']==[]
 sites={}
 for f in rows[1]['findings']:
  if f['kind'] not in ['Refused','NotYet']:continue
  where=f['where'].replace(str(S/label)+'/','');file,line,column=where.rsplit(':',2);line=int(line)
  if label not in ['main','head1-mutant'] and file=='src/compiler/sys.ts' and line>=57:line+=2
  key=(file,line,int(column),f['kind'],f['reason'],f['phase']);sites[key]={'where':file+':'+str(line)+':'+column,'rawWhere':f['where'],'kind':f['kind'],'reason':f['reason'],'phase':f['phase']}
 sets[label]=sites;summaries[label]={'sites':len(sites),'units':len(rows[1]['units']),'statuses':dict(collections.Counter(x['status'] for x in rows[1]['units'])),'sitesLedger':list(sites.values())}
assert set(sets['main'])==set(sets['head1-mutant'])
a=sets['head1-candidate'];b=sets['main'];new1=[a[k] for k in sorted(set(a)-set(b))];gone1=[b[k] for k in sorted(set(b)-set(a))];assert len(new1)==1 and new1[0]['where']=='src/compiler/sys.ts:1501:13' and new1[0]['reason']=='reading setTimeout';assert sum(x['reason']=='an ambient function declaration' for x in gone1)==2
before=sets['head1-candidate'];after=sets['head2-candidate'];new2=[after[k] for k in sorted(set(after)-set(before))];gone2=[before[k] for k in sorted(set(before)-set(after))]
allowed={(1501,13),(1502,13),(1599,13),(1600,13),(490,9),(496,37),(497,25),(1519,36)}
unexpected=[x for x in new2 if tuple(map(int,x['where'].rsplit(':',2)[1:])) not in allowed]
assert any(x['where']=='src/compiler/sys.ts:384:79' for x in gone2);assert any(k[1:3]==(384,79) for k in sets['head2-mutant'])
result={'summaries':summaries,'head1':{'pass':True,'expected995spev':new1,'gone':gone1,'mutantRestoresEntireBaseline':True},'head2':{'pass':not unexpected,'new':new2,'gone':gone2,'unexpected':unexpected,'mutantRestores384':True,'stoppingReason':'Unlisted new site requires a further ruling; no head2 adapter, API sanction, shard regeneration or lane run was constructed.' if unexpected else None}}
(S/'head2-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'head1':result['head1'],'head2':result['head2']},indent=2))
