const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const out=__dirname,base='/workspace/after-wave-adapted/src/compiler';
const groups=JSON.parse(fs.readFileSync(path.join(out,'grouped-first-stops.json'),'utf8'));
const records=[];
for(const g of groups.filter(g=>g.new_group&&g.where!=='-')) {
 const m=g.where.match(/^(.*):(\d+):(\d+)$/);if(!m)throw Error(g.where);
 const file=path.join(base,m[1]),source=fs.readFileSync(file,'utf8'),lines=source.split('\n'),sf=ts.createSourceFile(file,source,ts.ScriptTarget.Latest,true);
 const position=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);let scopes=[],nodes=[];
 function visit(n){if(n.getStart(sf)<=position&&position<n.end){nodes.push(n);if(ts.isFunctionLike(n))scopes.push(n);ts.forEachChild(n,visit);}}visit(sf);
 const scope=scopes.at(-1),start=scope?sf.getLineAndCharacterOfPosition(scope.getStart(sf)).line+1:+m[2],end=scope?sf.getLineAndCharacterOfPosition(scope.end).line+1:+m[2];
 const whole=!!scope&&end-start+1<30,text=whole?source.slice(sf.getPositionOfLineAndCharacter(start-1,0),scope.end):lines[+m[2]-1];
 const name=m[1].replaceAll('/','_')+'-'+m[2]+'-'+m[3]+'.txt';fs.writeFileSync(path.join(out,name),text+'\n');
 records.push({where:g.where,kind:g.kind,reason:g.reason,file,name,scope_name:scope?.name?.getText(sf)||null,scope_start:start,scope_end:end,whole_function:whole,text,site_text:nodes.at(-1)?.getText(sf)||null,source_line:lines[+m[2]-1]});
}
fs.writeFileSync(path.join(out,'source-excerpts.json'),JSON.stringify(records,null,2)+'\n');
