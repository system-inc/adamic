// Original member declarations and read spans, with explicitly reduced carriers.
const fs = require('fs'), path = require('path'), crypto = require('crypto'), cp = require('child_process');
const pin = process.argv[2], ts = require(process.argv[3]);
const sourceSha = '050880ce59e30b356b686bd3144efe24f875ebc8';
if (cp.execFileSync('git', ['-C', pin, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== sourceSha) throw Error('wrong source pin');
const lane = path.dirname(__dirname);
const pairs = JSON.parse(fs.readFileSync(path.join(lane,'unknown-callable-pairs-ranked.json')));
const recipes = new Map([
 [285, ['', ['"ok"'], 'boolean', 'true']],
 [306, ['interface PropertyAssignment {readonly value:number;}\ninterface Expression {readonly value:number;}\ninterface Identifier {readonly kind:80;readonly value:number;}\ninterface PrivateIdentifier {readonly kind:81;readonly value:number;}\ninterface StringLiteral {readonly kind:11;readonly value:number;}\ninterface NoSubstitutionTemplateLiteral {readonly kind:15;readonly value:number;}\ninterface NumericLiteral {readonly kind:9;readonly value:number;}\ninterface ComputedPropertyName {readonly kind:168;readonly value:number;}\ninterface BigIntLiteral {readonly kind:10;readonly value:number;}\ntype PropertyName = Identifier | StringLiteral | NoSubstitutionTemplateLiteral | NumericLiteral | ComputedPropertyName | PrivateIdentifier | BigIntLiteral;', ['{value:3}', '{kind:80,value:4}', '{value:5}'], 'PropertyAssignment', '{value:7}']],
 [333, ['', ['"ok"'], 'void', '']],
 [354, ['interface JSDocText {readonly value:number;}', ['"ok"'], 'JSDocText', '{value:7}']],
 [360, ['interface CommaListExpression {readonly value:number;}\ninterface Expression {readonly value:number;}', ['{value:3}','[{value:4}]'], 'CommaListExpression', '{value:7}']],
 [366, ['interface PropertyAccessExpression {readonly value:number;}\ninterface Expression {readonly value:number;}\ninterface Identifier {readonly kind:80;readonly value:number;}\ninterface PrivateIdentifier {readonly kind:81;readonly value:number;}\ntype MemberName = Identifier | PrivateIdentifier;', ['{value:3}','{value:4}','{kind:80,value:5}'], 'PropertyAccessExpression', '{value:7}']],
 [369, ['interface TypeNode {readonly value:number;}', ['{value:3}'], 'TypeNode', '{value:7}']],
 [372, ['interface ProjectReference {readonly value:number;}', [], 'readonly ProjectReference[] | undefined', '[{value:7}]']],
 [375, ['interface TypeChecker {readonly value:number;}', [], 'TypeChecker', '{value:7}']],
 [396, ['', ['"ok"'], 'string | undefined', '"ok"']],
 [399, ['', ['"ok"'], 'boolean', 'true']],
 [402, ['', [], 'string', '"ok"']],
 [426, ['', ['"ok"'], 'boolean', 'true']],
 [477, ['type TokenFlags = number;', [], 'TokenFlags', '7']],
 [486, ['', ['"ok"'], 'void', '']],
 [489, ['', ['"ok"'], 'string | undefined', '"ok"']],
 [513, ['', ['"ok"'], 'boolean', 'true']],
]);
const files = new Map();
function source(file) {if(!files.has(file)) files.set(file,ts.createSourceFile(file,fs.readFileSync(path.join(pin,file),'utf8'),ts.ScriptTarget.Latest,true));return files.get(file);}
const nodes=new Map();function collect(n){if(ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n))nodes.set(n.name.text,[...(nodes.get(n.name.text)||[]),n]);ts.forEachChild(n,collect);}collect(source('src/compiler/types.ts'));collect(source('src/compiler/scanner.ts'));collect(source('src/compiler/sys.ts'));collect(source('src/compiler/watchUtilities.ts'));
function declaration(type,field,seen=new Set()){if(seen.has(type))return;seen.add(type);const list=nodes.get(type)||[];for(const n of list){const own=n.members?.find(m=>m.name?.getText(n.getSourceFile())===field);if(own)return own;}for(const n of list)for(const c of n.heritageClauses||[])for(const b of c.types){const found=declaration(b.expression.getText(n.getSourceFile()),field,seen);if(found)return found;}}
const rows=[];
for(const [rank,recipe] of recipes){
 const pair=pairs.find(p=>p.rank===rank);if(rank%3)throw Error('wrong share');
 const decl=declaration(pair.type,pair.field);if(!decl||!ts.isMethodSignature(decl)||decl.typeParameters?.length||decl.parameters.some(p=>p.dotDotDotToken))throw Error('unsupported original declaration '+rank);
 const sf=source(pair.witness.file);let read;function visit(n){if(ts.isPropertyAccessExpression(n)&&n.name.text===pair.field){const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));if(lc.line+1===pair.witness.line&&lc.character+1===pair.witness.column)read=n;}ts.forEachChild(n,visit);}visit(sf);if(!read)throw Error('original read unavailable '+rank);
 const receiver=read.expression.getText(sf);if(!/^[A-Za-z_$][\w$]*(\.[A-Za-z_$][\w$]*)*$/.test(receiver.replace(/\(\)$/, '')))throw Error('nonlocal receiver '+rank+': '+receiver);
 let [carriers,args,result,returned]=recipe;
 for(const alias of ["PropertyName","MemberName"])if(carriers.includes("type "+alias+" =")){const original=nodes.get(alias)[0];carriers=carriers.replace(new RegExp("type "+alias+" =[^;]+;"),original.getText(original.getSourceFile()).replace(/^export\s+/,""));}
 const params=decl.parameters.map(p=>p.getText(decl.getSourceFile())).join(', ');
 const expression=read.getText(sf)+'('+args.join(',')+')';
 let use;if(result==='void') use=expression+';console.log("done");';else if(result.includes('ProjectReference[]'))use='const result='+expression+';console.log(`${result===undefined?"missing":result.length}`);';else if(/^(boolean|string|TokenFlags)/.test(result))use='console.log(`${'+expression+'}`);';else use='console.log(`${'+expression+'.value}`);';
 const parts=receiver.replace(/\(\)$/, '').split('.');let bound=receiver.endsWith('()')?'()=>value as Target':'value as Target';for(const key of parts.slice(1).reverse())bound='{'+key+':'+bound+'}';
 const header='// Original declaration and member read; adjacent carriers and producer bodies reduced.\n'+carriers+'\ninterface Base {readonly '+pair.field+':unknown;}\ninterface Target {'+decl.getText(decl.getSourceFile())+'}\nfunction probe(value:Base):void {const '+parts[0]+'='+bound+';'+use+'}\n';
 const body=result==='void'?'{}':'('+returned+')';
 const good='('+params+'):'+result+'=>'+body;
 const wrongArity=decl.parameters.length===0?'(ignored:number):'+result+'=>'+body:'():'+result+'=>'+body;
 const directory='rank-'+rank;fs.mkdirSync(path.join(__dirname,directory),{recursive:true});
 for(const [name,producer] of [['good',good],['wrong-arity',wrongArity],['wrong-value','7']])fs.writeFileSync(path.join(__dirname,directory,name+'.a'),header+'probe({'+pair.field+':'+producer+'});\n');
 const declarationSource=decl.getSourceFile(),hash=file=>crypto.createHash('sha256').update(fs.readFileSync(path.join(pin,file))).digest('hex');
 const expected=decl.getText(declarationSource).replace(pair.field,'').replace(/\)\s*:/,') =>').replace(/;$/, '').replace(/TokenFlags/g,'number').replace(/encoding\?: string/, 'encoding?: string | undefined').trim();
 rows.push({...pair,directory,expected,read:read.getText(sf),declaration:decl.getText(declarationSource),declarationFile:declarationSource.fileName,declarationSha256:hash(declarationSource.fileName),fileSha256:hash(pair.witness.file),utf16Start:read.getStart(sf),utf16End:read.end,arity:decl.parameters.length,mutantArity:decl.parameters.length===0?1:0,stdout:result==='void'?'done\n':result.includes('ProjectReference[]')?'1\n':result==='boolean'?'true\n':/^string/.test(result)?'ok\n':'7\n',carriers});
}
fs.writeFileSync(path.join(__dirname,'witnesses.json'),JSON.stringify({sourceSha,basis:'original member declarations/read spans; reduced adjacent carriers and implementations; candidate counts',members:rows},null,2)+'\n');
console.log('Prepared '+rows.length+' pairs / '+rows.reduce((n,r)=>n+r.reads,0)+' candidate reads.');
