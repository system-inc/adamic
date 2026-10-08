const fs=require('fs'),path=require('path');
const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,'original-members.json')));
const carriers=JSON.parse(fs.readFileSync('/tmp/lane5-c-carrier-declarations.json'));
const rows=[];
for(const m of evidence.members){
 if(m.field!=='push'||!/^\w+\[\]$/.test(m.type)||!/^\w+(\.\w+)*\.push$/.test(m.read)||m.declarations.length!==1)continue;
 const name=m.type.slice(0,-2),d=carriers[name];
 if(!d||d.kind!=='InterfaceDeclaration')continue;
 const carrier='interface '+name+' {readonly value:number;'+(d.kindValue===undefined?'':'readonly kind:'+d.kindValue+';')+'}';
 const good='{value:1'+(d.kindValue===undefined?'':',kind:'+d.kindValue)+'}',bad='{value:"bad"'+(d.kindValue===undefined?'':',kind:'+d.kindValue)+'}';
 const parts=m.read.split('.');parts.pop();let owner='(value as Target).items';for(let i=parts.length-1;i>0;i--)owner='{'+parts[i]+':'+owner+'}';
 const prefix='// Original Array intrinsic member and receiver; adjacent payload interface reduced.\n'+carrier+'\ninterface OriginalMember<T> {'+m.declarations[0].text+'}\ninterface Base {readonly items:unknown;}\ninterface Target {readonly items:'+m.type+';}\nfunction probe(value:Base):void {const '+parts[0]+'='+owner+';const witnessResult='+m.read+'('+good+');console.log(String(witnessResult));}\n';
 const dir=path.join(__dirname,'families','rank-'+m.rank);fs.mkdirSync(dir,{recursive:true});
 fs.writeFileSync(path.join(dir,'good.a'),prefix+'const seed:'+name+'='+good+';probe({items:[seed]});\n');
 fs.writeFileSync(path.join(dir,'wrong-element.a'),prefix+'probe({items:['+bad+']});\n');
 rows.push({rank:m.rank,reads:m.reads,type:m.type,read:m.read,declaration:m.declarations[0].text,diagnostic:'UNPINNED'});
}
fs.writeFileSync(path.join(__dirname,'array-candidates.json'),JSON.stringify(rows,null,2)+'\n');console.log('Prepared '+rows.length+' original Array.push receiver families.');
