// Minimal allocation and release cost of one-unit string slicing versus numeric reads.
import { panic, programArguments, readTextFile } from 'adamic';
const args = programArguments();
const read = readTextFile(args[0] ?? panic('usage: stringUnitScan.ts <file> slice|numeric <rounds>'));
if(read.kind === 'Error') panic(read.message);
const numeric = args[1] === 'numeric';
const rounds = Number(args[2] ?? '1');
function scan(text: string, numeric: boolean): number {
    let result = 0;
    for(let index = 0; index < text.length; index++) {
        if(numeric) result += text.charCodeAt(index);
        else result += text.slice(index, index + 1).charCodeAt(0);
    }
    return result;
}
let checksum = 0;
for(let round = 0; round < rounds; round++) checksum += scan(read.text, numeric);
console.log(String(checksum));
