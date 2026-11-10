const fs=require('fs'),path=require('path');
const ts=require(path.resolve(process.argv[2]));
const root=process.argv[3],output=process.argv[4],result={};
function walk(dir){for(const d of fs.readdirSync(dir,{withFileTypes:true})){const p=path.join(dir,d.name);if(d.isDirectory())walk(p);else if(/\.(ts|a)$/.test(p)&&!p.endsWith('.d.ts'))read(p);}}
function read(p){const text=fs.readFileSync(p,'utf8'),file=ts.createSourceFile(p,text,ts.ScriptTarget.Latest,true),rows=[];
function visit(n){if(ts.isFunctionDeclaration(n)||ts.isMethodDeclaration(n)||ts.isFunctionExpression(n)||ts.isArrowFunction(n)){let name=n.name?.getText(file);if(!name&&n.parent?.name)name=n.parent.name.getText(file);const loc=file.getLineAndCharacterOfPosition(n.getStart(file));rows.push({name:name||'<anonymous>',start:Buffer.byteLength(text.slice(0,n.getStart(file))),end:Buffer.byteLength(text.slice(0,n.end)),line:loc.line+1,column:loc.character+1,has_body:!!n.body});}ts.forEachChild(n,visit);}visit(file);result[path.relative(root,p)]=rows;}
walk(path.join(root,'src'));fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
