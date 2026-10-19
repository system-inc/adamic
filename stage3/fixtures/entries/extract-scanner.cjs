// Extract the exact scanner allocation and entries statement; stock 6.0.3 supplies enum values.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
assert.equal(ts.version,'6.0.3');
const root=path.resolve(process.argv[2]),out=__dirname;
function file(name){const text=fs.readFileSync(path.join(root,name),'utf8');return ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);}
function find(tree,predicate){const found=[];function visit(n){if(predicate(n))found.push(n);ts.forEachChild(n,visit);}visit(tree);assert.equal(found.length,1);return found[0];}
function span(n){const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return {file:sf.fileName,line:p.line+1,start:n.getStart(sf),end:n.end,text:n.getText(sf)};}
const scanner=file('src/compiler/scanner.ts'),core=file('src/compiler/corePublic.ts');
const allocation=find(scanner,n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(scanner)==='textToKeywordObj'));
const entries=find(scanner,n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(scanner)==='textToKeyword'));
const mapLike=find(core,n=>ts.isInterfaceDeclaration(n)&&n.name.text==='MapLike');
const object=allocation.declarationList.declarations[0].initializer;
assert.ok(ts.isObjectLiteralExpression(object));
const constants=[];
for(const property of object.properties){assert.ok(ts.isPropertyAssignment(property)&&ts.isPropertyAccessExpression(property.initializer));const name=property.initializer.name.text;assert.equal(typeof ts.SyntaxKind[name],'number');constants.push({name,value:ts.SyntaxKind[name]});}
assert.equal(new Set(constants.map(x=>x.name)).size,constants.length);
const provenance={typescript:ts.version,commit:'050880ce59e30b356b686bd3144efe24f875ebc8',files:[scanner,core].map(f=>({file:f.fileName,sha256:crypto.createHash('sha256').update(fs.readFileSync(path.join(root,f.fileName))).digest('hex')})),spans:[mapLike,allocation,entries].map(span),constants,entries_count:constants.length};
const fileName='01_scanner_keywords.a',destination=path.join(out,fileName);
const previousHeader=fs.existsSync(destination)?fs.readFileSync(destination,'utf8').split('\n')[0]:'// a-check: checked';
const body=`${previousHeader}
// From TypeScript 6.0.3, src/compiler/corePublic.ts:13
// From TypeScript 6.0.3, src/compiler/scanner.ts:${span(allocation).line}
// From TypeScript 6.0.3, src/compiler/scanner.ts:${span(entries).line}
// Step 12: fresh const allocation, complete enumerable own data fields proven.
// SyntaxKind support constants retain stock 6.0.3's exact keyword numbers; no enum reflection.
const SyntaxKind = {
${constants.map(x=>'    '+x.name+': '+x.value+',').join('\n')}
} as const;
type KeywordSyntaxKind = (typeof SyntaxKind)[keyof typeof SyntaxKind];
${mapLike.getText(core)}
${allocation.getText(scanner)}
${entries.getText(scanner)}
console.log('count:' + textToKeyword.size);
textToKeyword.forEach((value, key) => console.log(key + ':' + value));
`.replace(/\r\n/g,'\n');
fs.writeFileSync(destination,body);
fs.writeFileSync(path.join(out,'scanner-source.json'),JSON.stringify(provenance,null,2)+'\n');
console.log('extracted exact scanner allocation and Object.entries statement:',constants.length,'keywords');
