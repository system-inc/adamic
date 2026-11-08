'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const dir = fs.mkdtempSync(path.join(os.tmpdir(),'record-cycles-validation-'));
try {
  fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});
  fs.mkdirSync(path.join(dir,'src/tsc'),{recursive:true});
  fs.writeFileSync(path.join(dir,'src/compiler/tsconfig.json'),JSON.stringify({compilerOptions:{strict:true,target:'ES2020',module:'commonjs',types:['node'],lib:['es2020']},include:['../tsc/tsc.ts']}));
  fs.symlinkSync(path.resolve(__dirname,'../../api/node_modules'),path.join(dir,'node_modules'),'dir');
  fs.writeFileSync(path.join(dir,'src/tsc/tsc.ts'),`
export {};
declare const environment: NodeJS.ProcessEnv;
type PlainPartial = Partial<{ x: number }>;
type PlainPick = Pick<{ x: number }, 'x'>;
const partial: PlainPartial = {};
const picked: PlainPick = { x: 1 };
type MixedJson = string | MixedRecord | MixedJson[];
interface MixedRecord { [key: string]: MixedJson | undefined; }
function recover(): MixedJson | undefined { return {}; }
type FiniteRecord = Record<'a' | 'b', number>;
const finite: FiniteRecord = { a: 1, b: 2 };
interface Holder { entries: Dict; }
interface Dict { [key: string]: Holder; }
interface SelfRecord { [key: string]: SelfRecord; }
interface Control { [key: string]: number; }
interface FunctionRecord { [key: string]: () => Holder; }
interface ViaMap { [key: string]: Map<string, MapHolder>; }
interface MapHolder { entries: ViaMap; }
interface ViaMapKey { [key: string]: Map<KeyHolder, string>; }
interface KeyHolder { entries: ViaMapKey; }
const empty: Dict = {};
const computed: Dict = { ['entry']: { entries: empty } };
const recursive: SelfRecord = {};
const control: Control = { one: 1 };
const fn: FunctionRecord = {};
`);
  const log = path.join(dir,'run.log'),fd=fs.openSync(log,'w');
  const child=spawnSync(process.execPath,['--stack-size=16384',path.join(__dirname,'census.cjs'),dir,path.join(dir,'counts.json')],{stdio:['ignore',fd,fd]});
  fs.closeSync(fd);
  console.log(`fixture child exit=${child.status}`);
  if(child.error)throw child.error;
  assert.equal(child.status,0,fs.readFileSync(log,'utf8'));
  const d=JSON.parse(fs.readFileSync(path.join(dir,'counts.json'),'utf8'));
  assert.deepEqual(d.diagnostics,[]);
  assert.equal(d.records.some(r=>r.name==='PlainPartial' || r.name==='PlainPick'),false);
  const row=n=>d.records.find(r=>r.name===n);
  assert.equal(row('MixedRecord').unsureLiteralCount,1);
  assert.equal(row('MixedRecord').unambiguousLiteralCount,0);
  assert.equal(row('FiniteRecord').status,'unsure');
  assert.equal(row('FiniteRecord').literalCount,1);
  assert.equal(row('Dict').status,'cycle');
  assert.equal(row('Dict').literalCount,2);
  assert.equal(row('Dict').sites.filter(s=>s.empty).length,1);
  assert.equal(row('Dict').sites.filter(s=>s.computed).length,1);
  assert.deepEqual(row('Dict').cycle.map(n=>n.type),['Dict','Holder','Dict']);
  assert.deepEqual(row('SelfRecord').cycle.map(n=>n.type),['SelfRecord','SelfRecord']);
  assert.equal(row('Control').ownershipCycle,false);
  assert.equal(row('Control').literalCount,1);
  assert.equal(row('FunctionRecord').status,'unsure');
  assert.equal(row('ViaMap').ownershipCycle,true);
  assert.equal(row('ViaMapKey').ownershipCycle,true);
  assert.equal(d.evaluationOrder['src/tsc/tsc.ts'],0);
  console.log('validation passed');
} finally {
  fs.rmSync(dir,{recursive:true,force:true});
}
