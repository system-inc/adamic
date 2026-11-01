'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process'),ts=require('typescript');
assert.equal(process.argv.length,6,'usage: node verify.cjs <adapter> <before tree> <after tree> <output>');
const [adapter,before,after,output]=process.argv.slice(2).map(p=>path.resolve(p));const {adapt,prerequisite,headers,arrayBefore,arrayAfter,templates}=require(adapter);
const relative='src/compiler/commandLineParser.ts',old=fs.readFileSync(path.join(before,relative),'utf8'),file=path.join(after,relative),text=fs.readFileSync(file,'utf8');
let expected=old;
for(const [,a,b] of headers){assert.equal(expected.split(a).length-1,1);expected=expected.replace(a,b);}
expected=expected.replace(arrayBefore,arrayAfter);
for(const t of templates)expected=expected.replace(t,t.replace('${formattedValue}','${String(formattedValue)}'));
assert.equal(text,expected);
const snapshot=(tree,ref)=>cp.execFileSync('git',['-C',tree,'rev-parse',ref],{encoding:'utf8'}).trim();
assert.equal(snapshot(before,'refs/stage3/99-correctness-axolotl-g-after'),snapshot(after,'refs/stage3/99-correctness-axolotl-h-before'));
function countKinds(s){const source=ts.createSourceFile('commandLineParser.ts',s,99,true);let casts=0,assertions=0;(function visit(n){if(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n))casts++;if(ts.isNonNullExpression(n))assertions++;ts.forEachChild(n,visit);})(source);return{casts,assertions};}
assert.deepEqual(countKinds(old),countKinds(text));
const stat=fs.statSync(file,{bigint:true}),second=adapt(after);assert.equal(second.edits,0);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);assert.equal(fs.readFileSync(file,'utf8'),text);
const refusals=[];
for(const [name,mutate] of [
 ['without-99g',s=>s.replace(prerequisite,prerequisite.replace('value === undefined','value !== undefined'))],
 ['one-header-only',s=>s.replace(headers[0][1],headers[0][2])],
 ['templates-only',s=>s.replaceAll('${formattedValue}','${String(formattedValue)}')],
 ['array-only',s=>s.replace(arrayBefore,arrayAfter)],
 ['wrong-return-type',s=>text.replace(headers[0][2],headers[0][2].replace('string | undefined','string | null'))],
 ['extra-scalar-call',s=>s.replace('return formatSingleValue(value, map);','return formatSingleValue(formatSingleValue(value, map), map);')],
 ['non-null-assertion',s=>s.replace('return formatSingleValue(value, map);','return formatSingleValue(value, map)!;')],
 ['callback-coercion',s=>text.replace(arrayAfter,arrayAfter.replace('formatSingleValue(v, map)','String(formatSingleValue(v, map) ?? "")'))],
]){const root=fs.mkdtempSync(path.join(path.dirname(output),'contract-'));const p=path.join(root,relative);fs.mkdirSync(path.dirname(p),{recursive:true});const input=mutate(old);assert.notEqual(input,old);fs.writeFileSync(p,input);const mtime=fs.statSync(p,{bigint:true}).mtimeNs;assert.throws(()=>adapt(root));assert.equal(fs.readFileSync(p,'utf8'),input);assert.equal(fs.statSync(p,{bigint:true}).mtimeNs,mtime);refusals.push(name);}
const opts={fileName:'commandLineParser.ts',compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments:false}};
const a=ts.transpileModule(old,opts).outputText,b=ts.transpileModule(text,opts).outputText;
let js=a.replace(arrayBefore,arrayAfter);for(const t of templates)js=js.replace(t,t.replace('${formattedValue}','${String(formattedValue)}'));assert.equal(b,js);
assert(fs.readFileSync(path.join(before,'built/local/typescript.d.ts')).equals(fs.readFileSync(path.join(after,'built/local/typescript.d.ts'))));
const result={sourceRegions:5,javascriptChanging:true,javascriptBeforeBytes:Buffer.byteLength(a),javascriptAfterBytes:Buffer.byteLength(b),secondRun:second,secondRunNoWrite:true,noNewCastsOrAssertions:countKinds(text),refusals,publicApiByteIdentical:true};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
