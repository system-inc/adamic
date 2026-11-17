'use strict';
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
const tree=path.resolve(process.argv[2]);
const names=new Set(['getProperties','injectClassNamedEvaluationHelperBlockIfMissing','transformNamedEvaluation']);
const roots=[], candidates=[], counts={files:0,tsFiles:0};
function walk(dir){for(const ent of fs.readdirSync(dir,{withFileTypes:true})){if(['.git','node_modules','built','cohere'].includes(ent.name))continue;const p=path.join(dir,ent.name);if(ent.isDirectory())walk(p);else {counts.files++;if(/\.[cm]?tsx?$/.test(p)){counts.tsFiles++;if(p.startsWith(tree+'/src/'))roots.push(p);else {const text=fs.readFileSync(p,'utf8');if([...names].some(n=>text.includes(n)))candidates.push(p);}}}}}
walk(tree);
const options={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true,noEmit:true,types:[],allowImportingTsExtensions:true};
const program=ts.createProgram([...roots,...candidates],options), checker=program.getTypeChecker();
const targetFiles=new Set(['src/compiler/transformers/utilities.ts','src/compiler/transformers/namedEvaluation.ts']);
function loc(n){const f=n.getSourceFile(),p=f.getLineAndCharacterOfPosition(n.getStart(f));return {file:path.relative(tree,f.fileName),line:p.line+1,column:p.character+1};}
const calls=[],references=[],owners=[];
function owner(d){return d && ts.isFunctionDeclaration(d)&& names.has(d.name?.text)&&targetFiles.has(path.relative(tree,d.getSourceFile().fileName));}
function typ(n){return checker.typeToString(checker.getTypeAtLocation(n),n,ts.TypeFormatFlags.NoTruncation);}
for(const f of program.getSourceFiles()){if(!f.fileName.startsWith(tree+'/'))continue;function visit(n){if(owner(n))owners.push({...loc(n),name:n.name.text,body:!!n.body,text:n.getText(f),return:checker.typeToString(checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(n)),n,ts.TypeFormatFlags.NoTruncation)});
if(ts.isCallExpression(n)||ts.isNewExpression(n)||ts.isTaggedTemplateExpression(n)){const sig=checker.getResolvedSignature(n),d=sig?.declaration;if(owner(d)){let p=n.parent;calls.push({...loc(n),name:d.name.text,declaration:loc(d),result:typ(n),arguments:(n.arguments||[]).map(a=>({text:a.getText(f),type:typ(a)})),parent:ts.SyntaxKind[p.kind],context:p.getText(f),statement:(()=>{let s=n;while(s.parent&&!ts.isStatement(s))s=s.parent;return s.getText(f)})()});}}
if(ts.isIdentifier(n)&&names.has(n.text)){let s=checker.getSymbolAtLocation(n);if(s?.flags&ts.SymbolFlags.Alias)s=checker.getAliasedSymbol(s);if(s?.declarations?.some(owner))references.push({...loc(n),name:n.text,parent:ts.SyntaxKind[n.parent.kind]});}
ts.forEachChild(n,visit);}visit(f);}
console.log(JSON.stringify({typescript:ts.version,tree,coverage:{...counts,sourceRoots:roots.length,testCandidateRoots:candidates.length,programFiles:program.getSourceFiles().length},owners,calls,references},null,2));
