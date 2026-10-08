const fs=require('fs'),path=require('path'),crypto=require('crypto');
const root=process.argv[2],ts=require(path.join(root,'lib/typescript.js'));
const home=path.resolve(__dirname,'..'),members=require('../original-members.json').members;
const normalized=s=>s.replace(/\s+/g,'');
const rows=require('./namespace-probes.json');
for(const row of rows){
 const member=members.find(m=>m.rank===row.rank),original=fs.readFileSync(path.join(root,member.witness.file),'utf8');
 if(member.rank%3!==2||crypto.createHash('sha256').update(original).digest('hex')!==member.fileSha256||original.slice(member.start,member.end)!==member.read)throw Error('original source drift');
 if(normalized(row.originalHeader)!==normalized(member.declarations[0].text))throw Error('original header drift');
 const text=fs.readFileSync(path.join(home,'families','rank-'+row.rank,'original-context.a'),'utf8');
 const source=ts.createSourceFile('fixture.a',text,ts.ScriptTarget.Latest,true);
 const fn=source.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name.text===member.field);
 if(!fn?.body||normalized(text.slice(fn.getStart(source),fn.body.getStart(source))+';')!==normalized(row.originalHeader.replace(/^export /,'')))throw Error('fixture header drift');
 let read=false;function visit(n){if(ts.isPropertyAccessExpression(n)&&n.getText(source)===member.read)read=true;ts.forEachChild(n,visit);}visit(source);
 if(!read)throw Error('fixture read drift');
}
const sets=require('./SET_HANDOFF_058635b9.json');
console.log('Verified '+rows.length+' original namespace headers and read witnesses.');
