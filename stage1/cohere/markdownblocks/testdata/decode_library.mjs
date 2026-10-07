// Expose and run the exact private decodeString function from the pinned fork in memory.
// Its implementation, constants and regex literals remain the original bundle's.
import fs from 'node:fs';
import vm from 'node:vm';
const source = fs.readFileSync(process.argv[2] + '/plugins/markdown.js', 'utf8');
const anchor = 'function Hu(';
const start = source.indexOf(anchor);
if (start < 0 || source.indexOf(anchor, start + 1) >= 0) throw new Error('pinned decodeString anchor changed');
const declaration = /^function ([A-Za-z_$][\w$]*)\(/.exec(source.slice(start));
const context = {module:{exports:{}}, exports:{}};
vm.runInNewContext(source.slice(0,start) + `globalThis.__adamicDecodeString = ${declaration[1]};` + source.slice(start), context);
if (typeof context.__adamicDecodeString !== 'function') throw new Error('missing original function');
const decode = text => text.replace(/\\(.)/g, (_,c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const encode = text => text.replace(/[\\\n\r\t]/g, c => c === '\\' ? '\\\\' : c === '\n' ? '\\n' : c === '\r' ? '\\r' : '\\t');
const output = [];
for (const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if (line === '') continue;
    output.push(encode(context.__adamicDecodeString(decode(line.slice(1)))));
}
process.stdout.write(output.join('\n') + '\n');
