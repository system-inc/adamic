import { panic } from 'adamic';
function read(text: string): string {
    return text === 'a' ? 'a' : panic('unknown');
}
console.log(read('a'));
