const fs=require('fs'),path=require('path');
const ts=require('../../api/node_modules/typescript');
const root='/workspace/scratch/ff5eddz-adapted';
const result={};
function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=path.join(dir,e.name);if(e.isDirectory())walk(p);else if(p.endsWith('.ts')){
 const text=fs.readFileSync(p,'utf8'),source=ts.createSourceFile(p,text,ts.ScriptTarget.Latest,true),decls=[];
 function visit(n){if(ts.isFunctionLike(n)&&n.body){const start=n.getStart(source),lc=source.getLineAndCharacterOfPosition(start);decls.push({start,end:n.end,name:n.name?n.name.getText(source):'<anonymous>',where:p+':'+(lc.line+1)+':'+(lc.character+1)});}ts.forEachChild(n,visit);}visit(source);
 result[p]={sha256:require('crypto').createHash('sha256').update(fs.readFileSync(p)).digest('hex'),declarations:decls};
}}}walk(root+'/src');
fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n');
