"""Generate deterministic parser and depth controls; Go supplies actual ASTs."""
from pathlib import Path
import json
here=Path(__file__).resolve().parent
sources=['','const value=1','<div/>','<div><span/></div>','<><span/></>','useState(0)','usea()','use1()','React.useState(0)','react.useState(0)','obj.useState()','use_state()','function A(){if(flag){return <a/>;}else{return <b/>;}}','function A(){switch(x){case 1:return <a/>;default:return 1;}}','function A(){for(const x of xs){return <a/>;}}','function A(){try{return <a/>;}catch(e){return 1;}}','function A(){return <a/>; function B(){useState(0)}}','foo(1,2,<a/>)','const x=<><A/>{useState(0)}</>','const x=()=>()=> <a/>','class A {method(){return <a/>}}','const x=useSomething?.()','(useState)(0)','obj["useState"](0)','use💡()']
rows=[{'Name':'source-'+str(i),'Source':s} for i,s in enumerate(sources)]
for depth in range(33):
 for terminal in ['<div/>','<><div/></>','useState(0)','foo()']:
  source='const x='+('(' * depth)+terminal+(')' * depth)+';'
  rows.append({'Name':f'boundary-{depth}-{terminal}','Source':source,'RootOnly':True,'Depths':[-1,0,1,18,19,20,21]})
rows.append({'Name':'nil','Nil':True})
(here/'witnesses.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
print('controls',len(rows))
