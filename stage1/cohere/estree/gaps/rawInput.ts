import { fileStatus, panic, programArguments, readTextFile, utf8Length } from 'adamic';
const path = programArguments()[0] ?? panic('missing file');
const contents = readTextFile(path);
const status = fileStatus(path);
if(contents.kind === 'Error' || status.kind === 'Error') {
    panic('input failed');
}
console.log(`${status.size}:${utf8Length(contents.text)}:${contents.text}`);
