import { panic, programArguments, readTextFile } from 'adamic';
import { answer } from './pipeline.ts';
function run(path: string): void {
    const file = readTextFile(path);
    if(file.kind === 'Error') {
        panic(file.message);
    }
    console.log(answer(path, file.text).slice(0, -1));
}
const args = programArguments();
const path = args[0] ?? panic('usage: main.ts <file.ts> | --manifest <file>');
if(path === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    for(const item of manifest.text.split('\n')) {
        if(item !== '') {
            run(item);
        }
    }
}
else {
    run(path);
}
