// Go's simple Unicode case mapping differs from JavaScript's full casing.
import { panic } from 'adamic';
import { identifierCaseKeys } from './identifierCaseKeys.ts';
import { identifierCaseValues } from './identifierCaseValues.ts';
export function identifier(source: string): string {
    const output: string[] = [];
    let space = false;
    for(const character of source) {
        const code = character.codePointAt(0) ?? panic('identifier character');
        if(code === 9 || code === 10 || code === 13 || code === 32) {
            space = output.length > 0;
            continue;
        }
        if(space) output.push(' ');
        space = false;
        let low = 0;
        let high = identifierCaseKeys.length;
        while(low < high) {
            const middle = Math.floor((low + high) / 2);
            if((identifierCaseKeys[middle] ?? panic('identifier case key')) < code) low = middle + 1;
            else high = middle;
        }
        output.push(
            identifierCaseKeys[low] === code
                ? String.fromCodePoint(identifierCaseValues[low] ?? panic('identifier case value'))
                : character,
        );
    }
    return output.join('');
}
