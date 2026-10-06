import { panic } from 'adamic';

function make(): string {
    let text = '';
    for(let i = 0; i < 200; i++) text += 'x';
    return text;
}
function shared(): void {
    const holders: string[] = [make()];
    const held = holders[0] ?? panic('missing shared text');
    holders.splice(0, 1);
    console.log(`${held.length} ${held.slice(0, 4)}`);
}
shared();
