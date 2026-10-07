// Original pinned fork's document printer on canonical documents, not the port's list code.
import fs from 'node:fs';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
const prettier = require(process.argv[2] + '/standalone.js');
if(prettier.version !== '3.9.6') throw new Error('expected pinned fork 3.9.6');
const b = prettier.doc.builders;
const decode = text => text.replace(/\\([nrt\\])/g, (_, c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const encode = text => text.replace(/[\\\n\r\t]/g, c => c === '\\' ? '\\\\' : c === '\n' ? '\\n' : c === '\r' ? '\\r' : '\\t');
let documents = [], groups = new Map();
for(const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if(line === '') continue;
    const fields = line.split('\t');
    if(fields[0] === 'R') {
        const result = prettier.doc.printer.printDocToString(documents[+fields[1]], { printWidth:120, tabWidth:4, useTabs:false }).formatted;
        process.stdout.write(encode((fields[2] === '1' ? '\ufeff' : '') + result) + '\n');
        documents = []; groups = new Map(); continue;
    }
    if(fields[0] !== 'D') throw new Error('canonical opcode');
    const kind = fields[1], text = decode(fields[2]), width = +fields[3], flags = +fields[4], groupID = +fields[5];
    const children = fields[6] === '' ? [] : fields[6].split(',').map(id => documents[+id]);
    let id;
    if(groupID >= 0) { if(!groups.has(groupID)) groups.set(groupID, Symbol()); id = groups.get(groupID); }
    let doc;
    switch(kind) {
        case 't': doc = text; break;
        case 'a': doc = children; break;
        case 'h': doc = { type:'line', soft:!!(flags&1), hard:!!(flags&2), literal:!!(flags&4) }; break;
        case 'i': doc = b.indent(children[0]); break;
        case 's': doc = b.align(text, children[0]); break;
        case 'w': doc = b.align(width, children[0]); break;
        case 'r': doc = b.markAsRoot(children[0]); break;
        case 'd': doc = b.dedentToRoot(children[0]); break;
        case 'g': doc = b.group(children[0], { id, shouldBreak:!!(flags&1), ...(flags&2 ? { expandedStates:children } : {}) }); break;
        case 'f': doc = b.fill(children); break;
        case 'b': doc = b.ifBreak(children[0],children[1],{ groupId:id }); break;
        case 'l': doc = b.label('fixture',children[0]); break;
        case 'p': doc = b.breakParent; break;
        case 'x': doc = b.trim; break;
        default: throw new Error('canonical document kind ' + kind);
    }
    documents.push(doc);
}
