// Appending to a shared slice must leave its owner unchanged.
import { programArguments } from 'adamic';
const start = Number(programArguments()[0] ?? '0');
function probe(start: number): void {
    const source = 'a'.repeat(128);
    let prefix = source.slice(start, start + 80);
    prefix += 'x';
    console.log(source.slice(80, 81));
    console.log(prefix.slice(80));
}
probe(start);
