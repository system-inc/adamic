const ts=require('typescript');
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');
const file=process.argv[2]+'/src/compiler/debug.ts',text=fs.readFileSync(file,'utf8');
const ast=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const names=['setAssertionLevel','getAssertionLevel','fail','assertIsDefined','checkDefined'];
const found={};
function visit(n){if(ts.isFunctionDeclaration(n)&&names.includes(n.name?.text)) {assert(!found[n.name.text]); found[n.name.text]=n;} ts.forEachChild(n,visit);} visit(ast);
assert.equal(Object.keys(found).length,names.length);
for(const name of ['assertIsDefined','checkDefined']) assert(!/shouldAssert|currentAssertionLevel/.test(found[name].getText(ast)));
assert.equal(found.setAssertionLevel.body.getText(ast).replace(/\s/g,''),'{currentAssertionLevel=level;}');
const source='let currentAssertionLevel = 0;\n'+names.map(n=>found[n].getText(ast).replace(/^export /,'')).join('\n');
const js=ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.None}}).outputText;
const context={};vm.createContext(context);vm.runInContext(js,context);
const rows=[];
for(const level of [0,1,2,3]) {context.setAssertionLevel(level);let error;try{context.checkDefined(undefined);}catch(e){error=e.message;}assert.equal(error,'Debug Failure.'); assert.equal(context.checkDefined('ok'),'ok');rows.push({level,undefinedThrows:error,definedReturns:'ok'});}
console.log(JSON.stringify({file,functions:names.map(name=>({name,line:ast.getLineAndCharacterOfPosition(found[name].getStart(ast)).line+1,text:found[name].getText(ast)})),rows},null,2));
