'use strict';
const ts=require('typescript'),fs=require('fs');if(ts.version!=='6.0.3')throw Error(ts.version);
const s='/workspace/scratch/sys-timer-slot';
function members(file){const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true),rows=[];function visit(n){if(ts.isInterfaceDeclaration(n)&&n.name.text==='System')for(const m of n.members)if(['setTimeout','clearTimeout'].includes(m.name?.text))rows.push({name:m.name.text,start:m.getStart(),end:m.end,line:sf.getLineAndCharacterOfPosition(m.getStart()).line+1,endLine:sf.getLineAndCharacterOfPosition(m.end).line+1,text:m.getText()});ts.forEachChild(n,visit);}visit(sf);if(rows.length!==2)throw Error('expected System timer pair '+file);return rows;}
const result={main:{public:members(s+'/main/built/local/typescript.d.ts'),internal:members(s+'/main/built/local/typescript.internal.d.ts'),reference:members(s+'/main/tests/baselines/reference/api/typescript.d.ts')}};
for(const label of ['C-a','C-b','C-noinfer-a','C-noinfer-b'])result[label]={public:members(s+'/'+label+'/built/local/typescript.d.ts'),internal:members(s+'/'+label+'/built/local/typescript.internal.d.ts')};
fs.writeFileSync(process.argv[2],JSON.stringify({typescript:ts.version,trees:result},null,2)+'\n');console.log('declaration spans retained');
