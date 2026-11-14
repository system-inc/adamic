"""Select literal {}[] targets and arrays whose element is a declared generic constraint."""
import collections,json,pathlib,re,sys
rows=json.loads(pathlib.Path(sys.argv[1]).read_text());selected=[];review=[]
def normalized(s):return re.sub(r'[\s()]','',s)
for row in rows:
 elements=row.get('target_array_elements',[]);constraints=[x['constraint'] for x in row.get('constraints',[]) if x['constraint'] is not None]
 literal=any(normalized(x)=='{}' for x in elements)
 same=any(normalized(e) in [normalized(c),normalized(c+' | undefined')] for e in elements for c in constraints)
 direct=bool(elements) and any(normalized(row['target'])==normalized(c) for c in constraints)
 if literal or same:
  row['selection']='literal {} array target' if literal else 'array of declared type parameter constraint';selected.append(row)
 elif direct:
  row['selection']='related direct array-valued constraint';review.append(row)
 elif elements and constraints:
  row['selection']='other array view, not the constraint';review.append(row)
selected.sort(key=lambda r:(r['callee'],r['file'],r['line'],r['column'],r['reason']))
result={'count':len(selected),'by_callee':dict(sorted(collections.Counter(r['callee'] for r in selected).items())),'literal_empty_object_array':sum(r['selection']=='literal {} array target' for r in selected),'other_constraint_arrays':sum(r['selection']!='literal {} array target' for r in selected),'findings':selected,'other_constrained_array_targets_for_review':review}
pathlib.Path(sys.argv[2]).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='findings'},indent=2))
