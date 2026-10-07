"""Count distinct cast sites for overlapping flow patterns and exact transitions."""
import collections,gc,gzip,json,pathlib,sys

def read(path):
 with gzip.open(path,'rt') if str(path).endswith('.gz') else open(path) as stream:return json.load(stream)

def patterns(sites):
 counts=collections.defaultdict(collections.Counter)
 for row in sites:
  if row['reason']!="flow the graph can't see":continue
  detail='\n'.join(row['detail'])
  for label, needles in {
   'function-value escape callback boundary':['function value may receive untracked callback arguments'],
   'indirect/callback/intrinsic result boundary':['indirect, callback or intrinsic result not modeled'],
   'unbounded callback identity':['callback target identity is unknown'],
   'generic callback identity':['generic callback implementation not instantiated','generic callback substitution not modeled'],
   'dynamic keys':['dynamic element key','dynamic or relative element key'],
   'constructors':['constructor allocation and initialization body not modeled'],
   'whole-array contracts':['array intrinsic and augmented-field certificates unavailable'],
  }.items():
   if any(needle in detail for needle in needles):counts[label][row['kind']]+=1
 return {label:dict(counts[label]) for label in counts}

before=read(sys.argv[1]);old={(r['file'],r['start'],r['end']):{key:r[key] for key in ('reason','kind','outcome')} for r in before['sites']};before_patterns=patterns(before['sites']);before_diagnostics=set(before['diagnostics']);del before;gc.collect()
after=read(sys.argv[2]);transitions=collections.Counter();changes=[]
for row in after['sites']:
 previous=old[(row['file'],row['start'],row['end'])];transitions[(previous['reason'],row['reason'],row['kind'])]+=1
 if previous['reason']!=row['reason']:changes.append({key:row[key] for key in ('file','line','column','kind')}|{'before':previous['reason'],'after':row['reason']})
assert before_diagnostics==set(after['diagnostics']),'checker inventory changed'
result=dict(measurement=after['measurement'],note='Distinct sites per pattern; patterns overlap. Indirect/callback/intrinsic is a mixed boundary, not proof that every such site uses a callback.',before_patterns=before_patterns,after_patterns=patterns(after['sites']),transitions=[dict(before=a,after=b,kind=k,sites=n) for (a,b,k),n in sorted(transitions.items())],changed_sites=changes,unchanged_diagnostics=len(before_diagnostics),conforms_if=[{key:row[key] for key in ('file','line','column','kind','fields')} for row in after['sites'] if row['outcome']=='conforms-if'])
pathlib.Path(sys.argv[3]).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({key:result[key] for key in ('before_patterns','after_patterns','transitions','unchanged_diagnostics')},indent=2))
