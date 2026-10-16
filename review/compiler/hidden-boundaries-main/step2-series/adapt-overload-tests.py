from pathlib import Path
import re

def endbrace(s,start):
 level=0;i=start;quote=None
 while i<len(s):
  c=s[i]
  if quote:
   if c=='\\' and quote!='`':i+=2;continue
   if c==quote:quote=None
  elif c in '\"\'`':quote=c
  elif s.startswith('//',i):i=s.find('\n',i);continue
  elif s.startswith('/*',i):i=s.index('*/',i)+2;continue
  elif c=='{':level+=1
  elif c=='}':
   level-=1
   if level==0:return i
  i+=1
 raise Exception('unbalanced')

def flat(s):
 while '\tt.Run(' in s:
  start=s.index('t.Run(');op=s.index('func(t *testing.T) {',start);brace=op+len('func(t *testing.T) ');end=endbrace(s,brace)
  assert s[end+1:end+2]==')'
  s=s[:start]+'func() '+s[brace:end+1]+'()'+s[end+2:]
 return s

def suffix(name):return ''.join(p[0].upper()+p[1:] for p in re.split('[^a-zA-Z0-9]+',name) if p)
def split(path,name,cases,injections,params=('selected',)):
 p=Path(path);s=p.read_text();a=s.index('func '+name+'(');brace=s.index('{',a);b=endbrace(s,brace)+1;old=s[a:b];helper='check'+name[4:]
 old=old.replace('func '+name+'(t *testing.T)', 'func '+helper+'(t *testing.T, '+', '.join(params)+' string)',1).replace('\n\tt.Parallel()','',1)
 old=old[:old.index('{')+1]+'\n\tt.Helper()'+old[old.index('{')+1:]
 for needle,condition in injections:
  assert needle in old,(path,needle);old=old.replace(needle,needle+'\n'+condition,1)
 old=flat(old)
 wrappers=''
 for case in cases:
  args=(case,) if isinstance(case,str) else case
  wrappers+='\n\nfunc '+name+''.join(suffix(x) for x in args)+'(t *testing.T) {\n\tt.Parallel()\n\t'+helper+'(t, '+', '.join('"'+x+'"' for x in args)+')\n}'
 s=s[:a]+old+wrappers+s[b:];p.write_text(s)
for file,test,cases,key in [
 ('internal/lower/overload_results_test.go','TestOverloadResultsFixtures',['transform','evaluate','binding','parameter','scalar'],'name'),
 ('internal/lower/overload_results_test.go','TestOverloadResultsLiarStops',['liar','evaluate-liar','parameter-liar'],'name'),
 ('internal/lower/overload_results_test.go','TestOverloadResultsRefuses',['optional result property','shared writable result','escaping checked overload','unserved generic domain','shared readonly result'],'probe.name'),
 ('internal/lower/overload_structural_test.go','TestOverloadStructuralRefuses',['Block factory covariance','factory literal covariance','factory writable invariance','factory field storage','factory result covariance','visitor domain liar','visitor input contravariance','mixed field storage','missing Block field','shared wider result','reverse readonly variance','writable variance'],'probe.name'),
 ('internal/lower/overload_values_test.go','TestOverloadValueStops',['mutable target','opaque storage','mixed parameter'],'name'),
 ('internal/lower/overload_visitors_test.go','TestOverloadVisitorPaths',['unguarded presence','fabricated node','stored callback','returned callback','intervening element effect'],'probe.name')]:
 # Table literals end just before the loop body; the following t.Run starts each piece.
 p=Path(file);s=p.read_text();a=s.index('func '+test+'(');brace=s.index('{',a);b=endbrace(s,brace)+1;part=s[a:b];needle=next(x for x in part.splitlines() if 't.Run(' in x)
 split(file,test,cases,[(needle,'\t\t\tif '+key+' != selected { return }')])
 # Selection must precede the closure call and its removed t.Run line.
 p=Path(file);s=p.read_text();s=s.replace('func() {\n\t\t\tif '+key+' != selected { return }','func() {\n\t\t\tif '+key+' != selected { return }');p.write_text(s)
# Callback's names are computed before t.Run.
file='internal/lower/overload_callback_test.go';needle='\t\tname := []string{"input", "result", "generic", "alias", "representation"}[index]'
split(file,'TestOverloadCallbackUnserved',['input','result','generic','alias','representation'],[(needle,'\t\tif name != selected { continue }')])
for file,test,keys,cases in [
 ('internal/oracle/overload_field_hatch_test.go','TestOverloadFieldHatch',('field','variant'),[('kind','valid'),('kind','liar'),('value','valid'),('value','liar'),('value','undefined')]),
 ('internal/oracle/overload_values_test.go','TestOverloadValues',('name','variant'),[(n,v) for n in ['returned','narrow','order','module'] for v in ['valid','liar']]),
 ('internal/oracle/overload_visitors_test.go','TestOverloadVisitors',('name','variant'),[(n,v) for n in ['original','helper','overloaded-helper'] for v in ['valid','liar']]+[('default','liar')])]:
 p=Path(file);s=p.read_text();a=s.index('func '+test+'(');brace=s.index('{',a);part=s[a:endbrace(s,brace)+1];needle=next(x for x in part.splitlines() if 't.Run(' in x)
 # Remove t.Run first, then insert a filter inside its closure to preserve all scopes.
 s=s[:a]+flat(part)+s[a+len(part):];p.write_text(s)
 split(file,test,cases,[('func() {','\t\t\t\tif '+keys[0]+' != selectedName || '+keys[1]+' != selectedVariant { return }')],params=('selectedName','selectedVariant'))
# All imported top-level tests are parallel. Existing adjacent tests are untouched.
files=[Path(x) for x in ['internal/lower/overload_results_test.go','internal/lower/overload_structural_test.go','internal/lower/overload_callback_test.go','internal/lower/overload_values_test.go','internal/lower/overload_visitors_test.go','internal/oracle/overload_results_test.go','internal/oracle/overload_structural_test.go','internal/oracle/overload_callback_test.go','internal/oracle/overload_field_hatch_test.go','internal/oracle/overload_values_test.go','internal/oracle/overload_visitors_test.go','internal/oracle/hidden_boundary_generic_tnode_test.go']]+list(Path('cmd/adamic').glob('overload_*checks_test.go'))
for p in files:
 s=p.read_text();s=re.sub(r'(func Test\w+\(t \*testing.T\) \{\n)(?!\tt.Parallel\(\))',r'\1\tt.Parallel()\n',s);p.write_text(s)
