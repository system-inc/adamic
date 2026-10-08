const fs=require('node:fs');
const path=require('node:path');
const crypto=require('node:crypto');
const ts=require(process.env.SCOUT_TYPESCRIPT);
const tree=process.argv[2], closure=process.argv[3];
if(ts.version!=='6.0.3') throw Error('wrong TypeScript');
function read(file){const text=fs.readFileSync(file,'utf8');return {file,text,ast:ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),sha256:crypto.createHash('sha256').update(text).digest('hex')};}
function nodes(src,predicate){const found=[];function visit(n){if(predicate(n)) found.push(n);ts.forEachChild(n,visit);}visit(src.ast);return found;}
function locate(src,n){const pos=n.getStart(src.ast);const lc=src.ast.getLineAndCharacterOfPosition(pos);return {file:src.file,line:lc.line+1,column:lc.character+1,utf16_start:pos,utf16_end:n.end,text:n.getText(src.ast)};}
const scanner=read(path.join(tree,'src/compiler/scanner.ts'));
const adapted=read(path.join(closure,'src/compiler/scanner.ts'));
function identifierReturns(src){const f=nodes(src,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getIdentifierToken');if(f.length!==1) throw Error('function count');return nodes({...src,ast:f[0]},n=>ts.isReturnStatement(n)).map(n=>locate(src,n));}
const upstream=identifierReturns(scanner), gathered=identifierReturns(adapted);
if(upstream.length!==2 || gathered.length!==2 || upstream[0].text!=='return token = keyword;' || upstream[1].text!=='return token = SyntaxKind.Identifier;' || gathered[0].line!==1269 || upstream[0].line!==1823) throw Error('source mapping changed');
const types=read(path.join(tree,'src/compiler/types.ts'));
const enums=nodes(types,n=>ts.isEnumDeclaration(n)&&n.name.text==='SyntaxKind');
const keywords=nodes(types,n=>ts.isTypeAliasDeclaration(n)&&n.name.text==='KeywordSyntaxKind');
if(enums.length!==1 || keywords.length!==1) throw Error('type count');
const copied=fs.readFileSync(path.join(__dirname,'identifier-full-enum.a'),'utf8');
for(const n of [...enums,...keywords]) if(!copied.includes(n.getText(types.ast).replace(/\r\n/g,'\n').replace(/^export /,''))) throw Error('full type copy mismatch');
const table=nodes(scanner,n=>ts.isVariableDeclaration(n)&&n.name.getText(scanner.ast)==='textToKeywordObj');
const init=nodes(scanner,n=>ts.isVariableDeclaration(n)&&n.name.getText(scanner.ast)==='textToKeyword');
const memberNames=new Set(keywords[0].type.types.map(n=>n.getText(types.ast)));
if(table.length!==1 || table[0].initializer.properties.length!==84 || !table[0].initializer.properties.every(n=>n.initializer && memberNames.has(n.initializer.getText(scanner.ast)))) throw Error('table member provenance');
const bindings=nodes(scanner,n=>ts.isVariableDeclaration(n)&&n.name.getText(scanner.ast)==='token'&&n.type?.getText(scanner.ast)==='SyntaxKind');
console.log(JSON.stringify({typescript_version:ts.version,upstream_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',scanner_sha256:scanner.sha256,types_sha256:types.sha256,upstream_returns:upstream,gathered_returns:gathered,enum:locate(types,enums[0]),keyword_union:{...locate(types,keywords[0]),members:keywords[0].type.types.length},table:table.map(n=>({...locate(scanner,n),entries:n.initializer.properties.length})),map: init.map(n=>locate(scanner,n)),token:bindings.map(n=>locate(scanner,n)),reduced_values:{Identifier:ts.SyntaxKind.Identifier,AbstractKeyword:ts.SyntaxKind.AbstractKeyword,BreakKeyword:ts.SyntaxKind.BreakKeyword}},null,2));
