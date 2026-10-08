import { panic, programArguments, readTextFile } from 'adamic';
import { readCheckpoint, writeCheckpoint, decodeCheckpoint, encodeCheckpoint, InputFacts } from '../../replay/index.ts';
import type { Checkpoint, SidecarRow } from '../../replay/index.ts';
import type { HIRFunction, PlaceInterface } from '../../core.ts';
import { readPlace } from '../../replay/index.ts';
import { dropManualMemoization } from './drop_manual_memoization.ts';
import { inlineIIFE } from './inline_iife.ts';
import { mergeConsecutiveBlocks } from './merge_consecutive_blocks.ts';
import { independentResult, encodeResult } from './independent_steps.ts';
import { Unit3Facts } from './facts.ts';
function run(before: Checkpoint): string {
 const graph = decodeCheckpoint(before); const fn = graph.arena.read(graph.root); const facts = new Unit3Facts(before); let result = '';
 const pass = before.pass.split(':')[0] ?? panic('missing pass name');
 const independent = independentResult(graph,before,facts);
 if(independent !== undefined) { result = independent; }
 else if(pass === 'drop_manual_memoization') { const r = dropManualMemoization(fn); result = `${r.recognised},${r.marked},${r.dependencies},${r.withoutDepsArray},${r.unextractableDeps},${r.notAnArrayLiteral},${r.notAnInlineFunction}`; }
 else if(pass === 'inline_iife' || pass === 'inline_iife_including_memo_callbacks') { result = String(inlineIIFE(graph.arena,fn,pass === 'inline_iife_including_memo_callbacks')); }
 else if(pass === 'merge_consecutive_blocks') { result = String(mergeConsecutiveBlocks(fn)); }
 else { panic(`unknown unit-3 pass ${pass}`); }
 return encodeResult(graph,before,facts,result);
}
const args = programArguments(); const input = readTextFile(args[1] ?? panic('missing input'));
if(input.kind === 'Error') { panic(input.message); }
if(args[0] === '--checkpoint') { console.log(run(readCheckpoint(input.text)).slice(0,-1)); }
else if(args[0] === '--census') {
 for(const line of input.text.split('\n')) { if(line === '') { continue; } const fields = line.split('\t'); const path = fields[1] ?? panic('missing before path'); const text = readTextFile(path); if(text.kind === 'Error') { panic(text.message); } const before = readCheckpoint(text.text); if(`${before.key}/${before.pass}` !== fields[0]) { panic('manifest key differs'); } console.log(`checkpoint\t${fields[0]}`); console.log(run(before).slice(0,-1)); }
} else { panic('unknown unit-3 command'); }
