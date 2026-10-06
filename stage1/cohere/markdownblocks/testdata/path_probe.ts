import { panic, programArguments, readTextFile } from 'adamic';
import { AstArena } from '../astArena.ts';
import { AstPath, pathIndex, pathProperty } from '../astPath.ts';
import { readAstNode } from '../astProtocol.ts';
import { encode } from '../codec.ts';
import { observePath, pathSnapshot } from '../pathObserve.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: path_probe.ts <AST cases>'));
if(input.kind === 'Error') panic(input.message);
let arena = new AstArena();

const output: string[] = [];
for(const line of input.text.split('\n')) {
    const fields = line.split('\t');
    if(fields[0] === 'A') {
        arena = new AstArena();
    }
    else if(fields[0] === 'N') readAstNode(arena, fields);
    else if(fields[0] === 'F') {
        const root = 0;
        const path = new AstPath(arena, root);
        const rows: string[] = [];
        const next: number[] = [-1];
        const saved: number[] = [1];
        rows.push(observePath(path));
        while(next.length > 0) {
            const frame = next.length - 1;
            const id = path.node();
            const empty: number[] = [];
            const children = id < 0 ? empty : arena.node(id).children;
            const index = (next[frame] ?? panic('path frame')) + 1;
            next[frame] = index;
            if(index >= children.length) {
                path.truncate(saved.pop() ?? panic('path length'));
                next.pop();
                continue;
            }
            const before = path.stack.length;
            path.enter([pathProperty('children'), pathIndex(index)]);
            saved.push(before);
            next.push(-1);
            rows.push(arena.node(path.node()).children.length > 0 ? observePath(path) : pathSnapshot(path));
        }
        output.push(encode(rows.join('\n')));
    }
    else if(line !== '') panic('unknown path fixture');
}
console.log(output.join('\n'));
