// Expose and run the exact private splitText function from the pinned fork in memory.
// Its implementation, constants and regex literals remain the original bundle's.
import fs from 'node:fs';
import vm from 'node:vm';
const source = fs.readFileSync(process.argv[2] + '/plugins/markdown.js', 'utf8');
const split = '.split(/([\\t\\n ]+)/)';
const offset = source.indexOf(split);
if (offset < 0 || source.indexOf(split, offset + 1) >= 0) throw new Error('splitText anchor changed');
const start = source.lastIndexOf('function ', offset);
const declaration = /^function ([A-Za-z_$][\w$]*)\(/.exec(source.slice(start));
if (!declaration) throw new Error('splitText function declaration changed');
const context = {module:{exports:{}}, exports:{}};
vm.runInNewContext(source.slice(0,start) + `globalThis.__adamicSplitText = ${declaration[1]};` + source.slice(start), context);
if (typeof context.__adamicSplitText !== 'function') throw new Error('missing original function');
const decode = text => text.replace(/\\(.)/g, (_,c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const encode = text => text.replace(/[\\\n\r\t]/g, c => c === '\\' ? '\\\\' : c === '\n' ? '\\n' : c === '\r' ? '\\r' : '\\t');
const output = [];
for (const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if (line === '') continue;
    const fields = context.__adamicSplitText(decode(line.slice(1))).map(token => `${token.type}\t${token.kind ?? ''}\t${token.isCJ ? 1 : 0},${token.hasLeadingPunctuation ? 1 : 0},${token.hasTrailingPunctuation ? 1 : 0}\t${encode(token.value)}`);
    output.push(encode(fields.join('\n')));
}
process.stdout.write(output.join('\n') + '\n');
