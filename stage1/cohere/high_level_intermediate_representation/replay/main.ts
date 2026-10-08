import { panic, programArguments, readTextFile } from 'adamic';
import { readGraph, writeGraph, readCheckpoint, writeCheckpoint } from './checkpoint.ts';
import { encodeCheckpoint } from './encode.ts';
import { decodeCheckpoint } from './facts.ts';
import { dump } from '../dump.ts';
const args = programArguments();
const input = readTextFile(args[1] ?? panic('missing input'));
if(input.kind === 'Error') { panic(input.message); }
if(args[0] === '--graph') { console.log(writeGraph(readGraph(input.text)).slice(0,-1)); }
else if(args[0] === '--decode') { console.log(dump(decodeCheckpoint(readCheckpoint(input.text))).slice(0,-1)); }
else if(args[0] === '--checkpoint') { console.log(writeCheckpoint(readCheckpoint(input.text)).slice(0,-1)); }
else if(args[0] === '--census') {
    for(const line of input.text.split('\n')) {
        if(line === '') { continue; }
        const fields = line.split('\t'); const text = readTextFile(fields[1] ?? panic('missing checkpoint path'));
        if(text.kind === 'Error') { panic(text.message); }
        const checkpoint = readCheckpoint(text.text);
        if(checkpoint.key !== fields[0]) { panic('checkpoint key differs from manifest'); }
        const graph = writeGraph(checkpoint.graph);
        const decoded = decodeCheckpoint(checkpoint);
        const extraIdentities = checkpoint.identities.filter((identity) => identity.kind !== 'block');
        const extras = checkpoint.sidecars.filter((row) => !row.namespace.startsWith('input.'));
        if(writeCheckpoint(encodeCheckpoint(decoded,checkpoint,checkpoint.pass,extras,extraIdentities)) !== text.text) { panic(`arena checkpoint differs: ${checkpoint.key}`); }
        if(dump(decoded) !== graph) { panic(`decoded HIR differs: ${checkpoint.key}`); }
        if(writeGraph(readGraph(graph)) !== graph) { panic('graph roundtrip differs'); }
        console.log(`checkpoint\t${checkpoint.key}`); console.log(writeCheckpoint(checkpoint).slice(0,-1));
    }
} else { panic('unknown replay command'); }
