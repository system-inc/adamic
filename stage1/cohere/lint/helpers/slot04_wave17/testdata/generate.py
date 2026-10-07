"""Regenerate bounded color/length controls from the pinned Go color table."""
from pathlib import Path
import itertools, json, re
here=Path(__file__).resolve().parent
root=here.parents[5]
source=(root/'cohere/internal/lint/rules/tailwind/collapse/is_color.go').read_text()
names=re.findall(r'"([a-z]+)": true',source)
rows=[]
for n in names:
 for v in [n,n.upper(),n.title(),' '+n,n+' ',n+'x',n+'\n']:rows.append(v)
functions=['rgb(','rgba(','hsl(','hsla(','hwb(','color(','lab(','lch(','oklab(','oklch(','light-dark(','color-mix(','--alpha(','--spacing(']
for fn in functions:
 for v in [fn,fn.upper(),fn+'nonsense)',fn+'\n',fn[:-1],'x'+fn,' '+fn,'\n'+fn,fn+'💡']:rows.append(v)
 for i,c in enumerate(fn):
  if c.isalpha():
   for u in ['ſ','K','İ','ı','é','💡']:rows.append(fn[:i]+u+fn[i+1:])
for number,unit in itertools.product(['','0','1','.5','5.','+1','-1','1e','1e3','1e-3','NaN','∞','９','1 ',' 1'],['cm','mm','Q','q','px','PX','em','rem','cqw','cqmax','bogus','']):rows.append(number+unit)
rows+=['','#','#zzzz','##','x#fff','var(--color-red)','calc(1px)','xcalc(1px)','CALC(1px)','max(','rem(','min(1,2)','İndigo','darKblue','blacK','ſilver','\x00','\u2028','\u2029']
(here/'witnesses.json').write_text(json.dumps([{'Source':v} for v in sorted(set(rows))]+[{'Name':'all-named'}],ensure_ascii=True,indent=2)+'\n')
print(len(names),'named colors;',len(set(rows)),'controls; actual Go table replay added at runtime')
