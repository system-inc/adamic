// Research-only Node walk benchmark. Same unchanged source parser as native.
import { readFileSync } from 'node:fs';
import { availableParallelism, cpus } from 'node:os';
import { Parser } from '../parser/parser.ts';
import { countTree } from '../parser/nodes.ts';
const source = readFileSync(process.argv[2], 'utf8');
const flat = readFileSync(process.argv[3]);
const repeats = Number(process.argv[4] ?? 100);
// Warm module/parser allocation paths before taking a V8 retained-heap delta.
const warm = new Parser('x;');
countTree(warm.nodes, warm.file());
global.gc?.();
const before = process.memoryUsage().heapUsed;
const parser = new Parser(source, process.argv[2]);
const root = parser.file();
global.gc?.();
const after = process.memoryUsage().heapUsed;
const view = new DataView(flat.buffer, flat.byteOffset, flat.byteLength);
const n = view.getUint32(12, true);
const start = new Uint32Array(flat.buffer, flat.byteOffset + 40 + n*6*4, n);
const count = new Uint32Array(flat.buffer, flat.byteOffset + 40 + n*7*4, n);
const edges = new Uint32Array(flat.buffer, flat.byteOffset + 40 + n*48, view.getUint32(16, true));
function walk(i) { let answer = 1; for(let j = 0; j < count[i]; j++) answer += walk(edges[start[i]+j]); return answer; }
const current = () => countTree(parser.nodes, root);
const mapped = () => walk(view.getUint32(32, true));
const expected = current();
if(mapped() !== expected) throw Error('count parity');
const rows = [];
for(let round = 0; round < 5; round++) {
    for(const [name, run] of round % 2 ? [['flat', mapped], ['current', current]] : [['current', current], ['flat', mapped]]) {
        const begin = performance.now(); let answer = 0;
        for(let j = 0; j < repeats; j++) answer += run();
        const ms = performance.now() - begin;
        if(answer !== repeats * expected) throw Error('count mutant');
        rows.push({round, name, ms, answer});
    }
}
console.log(JSON.stringify({node:process.version, cpu:cpus()[0].model, available_cores:availableParallelism(), host_cores:cpus().length, repeats, reachable_visits:expected, table_nodes:parser.nodes.length, retained_heap_delta:after-before, retained_bytes_per_table_node:(after-before)/parser.nodes.length, flat_bytes:flat.length, rows}));
