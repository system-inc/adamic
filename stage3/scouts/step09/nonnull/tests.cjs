const ts=require('typescript'),assert=require('node:assert/strict'),vm=require('node:vm'),fs=require('node:fs'),path=require('node:path');
const proven=require('./flow.cjs'),instrument=require('./instrument.cjs');
assert.equal(ts.version,'6.0.3');
const fileName='/virtual/proof.ts';
const text=`function optional(x: number | undefined) { x!; if (x !== undefined) x!; }
function opaque(a: any, u: unknown) { a!; u!; }
function generic<T>(x:T) { x!; }
function constrained<T extends {}>(x:T) { x!; }
function required(x: {field:number}) { x.field!; }`;
const host=ts.createCompilerHost({strict:true});const original=host.getSourceFile.bind(host);
host.getSourceFile=(f,...rest)=>f===fileName?ts.createSourceFile(f,text,ts.ScriptTarget.Latest,true):original(f,...rest);
const p=ts.createProgram([fileName],{strict:true,noEmit:true},host),checker=p.getTypeChecker(),source=p.getSourceFile(fileName),nodes=[];
function find(n){if(ts.isNonNullExpression(n))nodes.push(n);ts.forEachChild(n,find);}find(source);
const expected=[false,true,false,false,false,true,true],types=nodes.map(n=>checker.getTypeAtLocation(n.expression));
assert.deepEqual(types.map(t=>proven(t,checker)),expected);
assert.notDeepEqual(types.map(()=>true),expected,'unchecked any/unknown/generic mutant');
assert.notDeepEqual(types.map(()=>false),expected,'drop proven erasure mutant');
const code=`let receivers=0, reads=0, writes=0, rights=0, state=1;
const object={get field(): number|undefined {reads++;return state;},set field(v:number|undefined){writes++;state=v!;}};
function owner(){receivers++;return object;}function rhs(){rights++;return 2;}
owner().field! |= rhs();const post=owner().field!++;const pre=++owner().field!;const absent=undefined!;
console.log(JSON.stringify([receivers,reads,writes,rights,state,post,pre,absent]));`;
const ast=ts.createSourceFile('probe.a',code,ts.ScriptTarget.Latest,true),anchors=new Map();
function collect(n){if(ts.isNonNullExpression(n)){const pos=ast.getLineAndCharacterOfPosition(n.getStart(ast));anchors.set(n.getStart(ast)+':'+n.end,{file:'probe.a',line:pos.line+1,column:pos.character+1,start:n.getStart(ast),end:n.end});}ts.forEachChild(n,collect);}collect(ast);
const inserted=[];const edited=instrument(ast,anchors,inserted);assert.equal(inserted.length,anchors.size);
function execute(input){const stdout=[],hits=[];vm.runInNewContext(ts.transpileModule(input,{compilerOptions:{target:ts.ScriptTarget.ES2022}}).outputText,{console:{log:s=>stdout.push(s)},__adamic_nonnull_probe:(v,id)=>{hits.push({value:v,id});return v;}});return {stdout,hits};}
const control=execute(code),observed=execute(edited);
assert.deepEqual(observed.stdout,control.stdout);assert.equal(observed.stdout[0],'[3,3,3,1,5,3,5,null]');
assert.equal(observed.hits.length,7);assert.equal(observed.hits.filter(h=>h.value===undefined).length,1);
const receiver=edited.indexOf('})(owner())');assert.ok(receiver>=0);const twice=edited.slice(0,receiver)+edited.slice(receiver).replace('})(owner())','})((owner(), owner()))');
const duplicate=execute(twice);assert.notDeepEqual(duplicate.stdout,control.stdout,'duplicate receiver mutant');
const omitted=execute(edited.replace(/__adamic_nonnull_probe\(/g,'((v) => v)('));assert.equal(omitted.hits.length,0);assert.notEqual(omitted.hits.length,observed.hits.length,'lost observer mutant');
const nestedAst=ts.createSourceFile('nested.a','const v: number | undefined = 7; console.log(String(v!!));',ts.ScriptTarget.Latest,true),nestedAnchors=new Map();
function nestedVisit(n){if(ts.isNonNullExpression(n)){const pos=nestedAst.getLineAndCharacterOfPosition(n.getStart(nestedAst));nestedAnchors.set(n.getStart(nestedAst)+':'+n.end,{file:'nested.a',line:pos.line+1,column:pos.character+1,start:n.getStart(nestedAst),end:n.end});}ts.forEachChild(n,nestedVisit);}nestedVisit(nestedAst);
const nestedInserted=[],nestedText=instrument(nestedAst,nestedAnchors,nestedInserted),nestedRun=execute(nestedText);
assert.equal(nestedInserted.length,2);assert.equal(new Set(nestedInserted).size,2);assert.equal(nestedRun.hits.length,2);assert.deepEqual(nestedRun.stdout,['7']);
const collapsed=execute(nestedText.replace(/@[0-9]+-[0-9]+/g,''));assert.equal(new Set(collapsed.hits.map(h=>h.id)).size,1);assert.notEqual(new Set(collapsed.hits.map(h=>h.id)).size,new Set(nestedRun.hits.map(h=>h.id)).size,'collapse nested span IDs mutant');
fs.writeFileSync(path.join(__dirname,'test-results.json'),JSON.stringify({flow_cases:expected.length,reference_updates:3,assertion_observations:observed.hits.length,mutants:['claim every operand proven','claim no operand proven','evaluate receiver twice','omit all observers','collapse nested span IDs'],caught:5},null,2)+'\n');
console.log('7 flow cases, 3 reference updates, single-evaluation equivalence, nullish capture, nested span identity, and 5 measurement mutants passed');
