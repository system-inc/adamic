// Node-only format reference. readFileSync allocates the backing buffer;
// columns are views, not deserialized arrays. This does not claim Node mmap.
import { readFileSync } from 'node:fs';
const b = readFileSync(process.argv[2]);
const h = new DataView(b.buffer, b.byteOffset, b.byteLength);
if(b.subarray(0, 8).toString('hex') !== '4144464c41540001') throw Error('magic');
const [size, n, children, roots, strings, units, root, reserved] = Array.from({length: 8}, (_, i) => h.getUint32(8 + 4*i, true));
if(size !== b.length || reserved !== 0) throw Error('header');
const cols = Array.from({length: 12}, (_, i) => new Uint32Array(b.buffer, b.byteOffset + 40 + i*n*4, n));
let off = 40 + 48*n;
const edges = new Uint32Array(b.buffer, b.byteOffset + off, children); off += 4*children;
const expr = new Uint32Array(b.buffer, b.byteOffset + off, roots); off += 4*roots;
const ranges = new Uint32Array(b.buffer, b.byteOffset + off, 2*strings); off += 8*strings;
const pool = new DataView(b.buffer, b.byteOffset + off, 2*units);
function str(id, escaped = false) {
    let out = '';
    for(let i = ranges[2*id]; i < ranges[2*id] + ranges[2*id+1]; i++) {
        const u = pool.getUint16(2*i, true);
        out += escaped && (u < 32 || u > 126 || u === 92) ? `\\u${u.toString(16).padStart(4, '0')}` : String.fromCharCode(u);
    }
    return out;
}
console.log(`case ${process.argv[3] ?? 0} ${root} ${Array.from(expr).join(',')}`);
for(let i = 0; i < n; i++) {
    console.log([str(cols[0][i]), cols[1][i], cols[2][i], cols[3][i] & 32, cols[4][i], cols[5][i] | 0, (cols[3][i] >>> 30) & 1, (cols[3][i] >>> 29) & 1, str(cols[10][i]), str(cols[8][i], true), str(cols[9][i], true), str(cols[11][i]), Array.from(edges.subarray(cols[6][i], cols[6][i] + cols[7][i])).join(',')].join('\t'));
}
