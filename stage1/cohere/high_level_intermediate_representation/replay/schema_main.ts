import { panic, programArguments, readTextFile } from 'adamic';
import { readCheckpoint, writeCheckpoint, decodeCheckpoint, encodeCheckpoint, finalizeHIR, CloneFunction } from './index.ts';
import { dump } from '../dump.ts';
const args = programArguments(); const input = readTextFile(args[1] ?? panic('missing checkpoint'));
if(input.kind === 'Error') { panic(input.message); }
const checkpoint = readCheckpoint(input.text); const graph = decodeCheckpoint(checkpoint);
if(args[0] === '--roundtrip') {
 const extras = checkpoint.sidecars.filter((row) => !row.namespace.startsWith('input.'));
 console.log(writeCheckpoint(encodeCheckpoint(graph,checkpoint,checkpoint.pass,extras,checkpoint.identities.filter((id) => id.kind !== 'block'))).slice(0,-1));
} else if(args[0] === '--clone') { console.log(dump(CloneFunction(graph)).slice(0,-1)); }
else if(args[0] === '--finalize') { finalizeHIR(graph.arena,graph.root); console.log(dump(graph).slice(0,-1)); }
else { panic('unknown schema command'); }
