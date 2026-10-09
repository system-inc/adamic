const fs=require('fs'),ts=require('/tmp/parser-fixed-feature-scratch/stage3/api/node_modules/typescript');
const [file,line,column]=process.argv.slice(2);const text=fs.readFileSync(file,'utf8');const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const pos=sf.getPositionOfLineAndCharacter(+line-1,+column-1);let candidates=[];
function walk(n){if(n.pos<=pos&&pos<n.end){if(ts.isFunctionDeclaration(n)||ts.isMethodDeclaration(n)||ts.isFunctionExpression(n)||ts.isArrowFunction(n))candidates.push(n);ts.forEachChild(n,walk);}}
walk(sf);const n=candidates.at(-1);if(!n)throw Error('No function stub for diagnostic');
let edits=[];function predicate(t){if(ts.isTypePredicateNode(t))edits.push({start:t.getStart(sf),end:t.end,replacement:'boolean'});else ts.forEachChild(t,predicate);}
for(const p of n.parameters)if(p.type)predicate(p.type);
if(n.body)edits.push({start:n.body.getStart(sf),end:n.body.end,replacement:'{ throw "scratch-only discovery stub";'+ '\n'.repeat((text.slice(n.body.getStart(sf),n.body.end).match(/\n/g)||[]).length)+'}'});
else edits.push({start:n.getStart(sf),end:n.end,replacement:'\n'.repeat((text.slice(n.getStart(sf),n.end).match(/\n/g)||[]).length)});
let output=text;for(const e of edits.sort((a,b)=>b.start-a.start))output=output.slice(0,e.start)+e.replacement+output.slice(e.end);
if(output===text)throw Error('Stub does not change source');fs.writeFileSync(file,output);console.log(JSON.stringify({file,line:+line,column:+column,name:n.name?.getText(sf),edits}));
