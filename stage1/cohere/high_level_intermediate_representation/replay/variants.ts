import { HIRArena, ConstructedHIR } from '../core.ts';
import { dump } from '../dump.ts';
import type { ManualMemoDependencyInterface, DependencyPathEntryInterface } from '../core.ts';
import { CloneFunction } from '../clone.ts';
import { panic } from 'adamic';
const arena = new HIRArena(); const root = arena.create('f'); const fn = arena.read(root); const p = fn.returns;
fn.emit(p,{kind: 'DeclareContext',lvalue: p,declarationKind: 0},0,0);
fn.emit(p,{kind: 'StartMemoize',manualMemoId: 1,deps: undefined},0,0);
const emptyDeps: ManualMemoDependencyInterface[] = []; const emptyPath: DependencyPathEntryInterface[] = [];
fn.emit(p,{kind: 'StartMemoize',manualMemoId: 2,deps: emptyDeps},0,0);
fn.emit(p,{kind: 'StartMemoize',manualMemoId: 3,deps: [{root: {isGlobal: false,place: p,name: ''},path: [{property: 'x',optional: true}]},{root: {isGlobal: true,place: p,name: 'React'},path: emptyPath}]},0,0);
fn.emit(p,{kind: 'FinishMemoize',manualMemoId: 3,value: p,pruned: false},0,0);
fn.emit(p,{kind: 'FinishMemoize',manualMemoId: 2,value: p,pruned: true},0,0);
const graph = new ConstructedHIR(arena,root); const original = dump(graph); const copy = CloneFunction(graph);
if(dump(copy) !== original) { panic('new instruction clone changes graph'); }
for(const instruction of copy.arena.read(copy.root).instructions) { const value = instruction.value; if(value.kind === 'StartMemoize' && value.deps !== undefined) { value.deps.pop(); } }
if(dump(graph) !== original) { panic('memo dependencies alias in clone'); }
for(const line of original.split('\n')) { if(line.startsWith('orphan instruction ')) { console.log(line.slice(7)); } }
