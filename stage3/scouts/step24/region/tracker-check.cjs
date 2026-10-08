const assert=require('node:assert/strict'),path=require('node:path');
const {makeTracker}=require(process.env.REGION_TRACKER||'./tracker.cjs');
const tracker=globalThis.__region=makeTracker();if(process.env.REGION_FORCE_PARSER_FILE==='1'){const enter=tracker.enter;tracker.enter=(site,parser)=>enter(site,parser||site.startsWith('src/compiler/parser.ts:'));}
const ts=require(path.resolve(process.argv[2],'instrumented.cjs'));tracker.end([],ts);
tracker.begin();const file=ts.createSourceFile('root.ts','let answer = 42;',ts.ScriptTarget.Latest,true);const root=tracker.end([file],ts);
assert(root.totals.created>1);assert.equal(root.totals.noParent,1);assert.equal(root.totals.notGraphReachable,0);assert.equal(root.totals.detachedExcludingSourceFileRoots,0);
tracker.begin();const node=ts.factory.createIdentifier('synthetic');const synth=tracker.end([],ts);assert.equal(synth.totals.synthetic,1);assert.equal(synth.totals.notGraphReachable,1);assert.equal(node.parent,undefined);
tracker.begin();ts.forEachChild(file,()=>{ts.factory.createIdentifier('walk-callback');return true;});const walking=tracker.end([file],ts);assert.equal(walking.totals.parser,0);assert.equal(walking.totals.synthetic,1);
// A missed seam must be caught by the independent constructor-entry audit.
tracker.begin();const record=tracker.record;tracker.record=node=>node;ts.factory.createIdentifier('mutant');tracker.record=record;
assert.throws(()=>tracker.end([],ts),/Allocation coverage mismatch/);
console.log('Root exemption and detached synthetic checks PASS; omitted allocation seam mutant caught by constructor audit.');
