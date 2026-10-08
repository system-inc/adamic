import { panic } from 'adamic';
import { HIRArena, replaceInstructionValue, InstructionIndex } from '../core.ts';
import { copyInstructionValueWithRemap, copyTerminalWithRemap } from '../clone.ts';
import { HIRFile, ForFunction, ForFunctionWithoutManualMemoization, MayNameManualMemoization } from '../cache.ts';
import { SymbolSnapshot } from '../symbol.ts';
const arena = new HIRArena(); const root = arena.create('f'); const fn = arena.read(root);
const first = fn.named('first',0,1); const declarations = fn.declarationIndices.length;
const second = fn.named('second',2,3,fn.identifier(first.identifier).declaration);
if(fn.declarationIndices.length !== declarations) { panic('existing declaration minted twice'); }
const copied = copyInstructionValueWithRemap(fn,fn,{kind: 'LoadLocal',place: first},(place) => ({identifier: second.identifier,effect: place.effect,reactive: place.reactive,start: place.start,end: place.end}),(reference) => reference);
if(copied.kind !== 'LoadLocal' || copied.place.identifier !== second.identifier) { panic('place copy ignores remap'); }
const id = fn.emit(first,{kind: 'Debugger'},0,1); replaceInstructionValue(fn,id,copied);
InstructionIndex.read(fn.instructionIndices,id);
if(fn.instruction(id).value.kind !== 'LoadLocal') { panic('value replacement failed'); }
const block = fn.newBlock('block');
const terminal = copyTerminalWithRemap({kind: 'Goto',block: fn.entry,variant: 0},(place) => place,(_old) => block.id);
if(terminal.block !== block.id) { panic('terminal copy ignores remap'); }
const symbols = new SymbolSnapshot('1\n0');
const source = 'function f() { return 1; }'; const file = new HIRFile(source,symbols);
const node = file.at(0,source.length); const intact = ForFunction(file,node);
let drops = 0; let inlines = 0;
const dropCount = (): number => drops; const inlineCount = (): number => inlines;
const drop = (): void => { drops++; }; const inline = (): number => { inlines++; return 0; };
if(ForFunctionWithoutManualMemoization(file,node,drop,inline) !== intact || dropCount() !== 0 || inlineCount() !== 0) { panic('no-memo cache entry did not share'); }
const memoSource = 'function f() { return "useMemo"; }'; const memo = new HIRFile(memoSource,symbols); const memoNode = memo.at(0,memoSource.length);
const rewritten = ForFunctionWithoutManualMemoization(memo,memoNode,drop,inline);
if(rewritten === undefined || rewritten !== ForFunctionWithoutManualMemoization(memo,memoNode,drop,inline) || dropCount() !== 1 || inlineCount() !== 1) { panic('memo cache re-executed passes'); }
const missing = new HIRFile(source,undefined);
if(ForFunctionWithoutManualMemoization(missing,missing.at(0,source.length),drop,inline) !== undefined) { panic('nil checker did not decline'); }
const escapedSource = 'function f() { return "use\\x4demo"; }'; const escaped = new HIRFile(escapedSource,symbols);
if(!MayNameManualMemoization(escaped,escaped.at(0,escapedSource.length))) { panic('cooked memo name was missed'); }
console.log('declaration/copy/rewrite/cache contracts pass');
