const ts=require(process.env.METHOD_VALUES_TYPESCRIPT || 'typescript');
const fs=require('fs'), path=require('path'), cp=require('node:child_process');
const repository=path.resolve(__dirname,'../../..');
const root=path.join(repository,'cohere/TypeScript/tsc/testdata/fixtures/compiler');
const files=[]; function walk(p){for(const d of fs.readdirSync(p,{withFileTypes:true})){const q=path.join(p,d.name);if(d.isDirectory())walk(q);else if(q.endsWith('.ts'))files.push(q)}} walk(root);
const program=ts.createProgram(files,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true});
const checker=program.getTypeChecker();
const bodies=new Map();
function usesThis(body){let reads=false;function visit(n){if(n.kind===ts.SyntaxKind.ThisKeyword)reads=true;if(n!==body && ts.isFunctionLike(n)&&!ts.isArrowFunction(n))return;ts.forEachChild(n,visit)}visit(body);return reads}
for(const file of program.getSourceFiles()){if(!file.fileName.startsWith(root))continue;function visit(n){if(ts.isFunctionLike(n)&&n.body&&n.name){const name=n.name.getText(file);const list=bodies.get(name)||[];const pos=file.getLineAndCharacterOfPosition(n.getStart(file));list.push({file:path.relative(root,file.fileName),line:pos.line+1,this:usesThis(n.body)});bodies.set(name,list)}ts.forEachChild(n,visit)}visit(file)}
const control=ts.createSourceFile('control.ts', 'function free() { return () => 1; } function lexical() { return () => this.value; } function separate() { return function () { return this.value; }; }', ts.ScriptTarget.ESNext, true);
const expected=[false,true,false];
for(let index=0;index<3;index++) if(usesThis(control.statements[index].body)!==expected[index]) throw Error('lexical this census control '+index+' failed');
const result=[];
const raw=cp.execFileSync('git',['show','d35a81d36fdafccf827bad0f572d311b2a0d4deb:stage3/refusal-table/raw.csv'],{cwd:repository,maxBuffer:16*1024*1024});
const rows=JSON.parse(cp.execFileSync('python3',['-c',"import csv,sys,json; print(json.dumps([r for r in csv.DictReader(sys.stdin) if r['reason'].startswith('a method read as a value')]))"],{input:raw}));
for(const row of rows){const file=program.getSourceFile(path.join(root,row.file));const position=file.getPositionOfLineAndCharacter(+row.line-1,+row.column-1);let access;function visit(n){if(ts.isPropertyAccessExpression(n)&&n.getStart(file)===position && n.name.text===row.reason.split("(")[1].split(" ")[0])access=n;ts.forEachChild(n,visit)}visit(file);if(!access)throw Error('site not found '+row.file+':'+row.line);const symbol=checker.getSymbolAtLocation(access.name);const declared=(symbol?.declarations||[]).filter(d=>d.body);const candidates=declared.length?declared.map(d=>({file:path.relative(root,d.getSourceFile().fileName),line:d.getSourceFile().getLineAndCharacterOfPosition(d.getStart()).line+1,this:usesThis(d.body)})):bodies.get(access.name.text)||[];const internal= /^(factory|parenthesizer)(\.|Rules\(\)\.)/.test(access.getText(file));
const classification=internal && candidates.length && !candidates.some(c=>c.this)?'this-free implementation family':declared.length && candidates.some(c=>c.this)?'this-reading body':'unresolved runtime origin';
result.push({classification,file:row.file,line:+row.line,column:+row.column,expression:access.getText(file),name:access.name.text,directBody:declared.length>0,candidates,status:candidates.length?(candidates.some(c=>c.this)?'this-reading candidate':'all named bodies this-free'):'no body found'})}
fs.writeFileSync(path.join(__dirname,'census.json'),JSON.stringify(result,null,2)+'\n');const counts={};for(const r of result)counts[r.classification]=(counts[r.classification]||0)+1;console.log(JSON.stringify(counts));console.log('direct bodies:',result.filter(r=>r.directBody).length);
