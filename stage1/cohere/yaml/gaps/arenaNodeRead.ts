// Same arena reads through an owning accessor or a direct array element.
import { panic, programArguments, readTextFile } from 'adamic';
class Entry {
    text: string;
    constructor(text: string) { this.text = text; }
}
class Entries {
    nodes: Entry[];
    constructor(text: string) { this.nodes = [new Entry(text)]; }
    node(index: number): Entry { return this.nodes[index] ?? panic('missing entry'); }
}
const args = programArguments();
const read = readTextFile(args[0] ?? panic('usage: arenaNodeRead.ts <file> accessor|direct <rounds>'));
if(read.kind === 'Error') panic(read.message);
const entries = new Entries(read.text);
const nodes = entries.nodes;
const direct = args[1] === 'direct';
const rounds = Number(args[2] ?? '1');
let checksum = 0;
for(let round = 0; round < rounds; round++) {
    for(let index = 0; index < read.text.length; index++) {
        if(direct) {
            const node = nodes[0] ?? panic('missing entry');
            checksum += node.text.length;
        }
        else {
            const node = entries.node(0);
            checksum += node.text.length;
        }
    }
}
console.log(String(checksum));
