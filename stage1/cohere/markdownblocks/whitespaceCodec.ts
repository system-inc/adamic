// Fixture transport only; production whitespace APIs take typed AST facts.
import { panic } from 'adamic';
import { decode } from './codec.ts';
import type { WhitespaceFrameInterface, WhitespaceTokenInterface, CJSpaceSampleInterface } from './whitespace.ts';
function number(value: string): number {
    const result = Number.parseInt(value, 10);
    if(!Number.isInteger(result)) panic('whitespace fixture integer');
    return result;
}
function token(fields: readonly string[], start: number): WhitespaceTokenInterface {
    const flags = number(fields[start + 4] ?? '');
    return {
        present: fields[start] === '1',
        type: fields[start + 1] ?? '',
        value: decode(fields[start + 2] ?? ''),
        kind: fields[start + 3] ?? '',
        cj: (flags & 1) !== 0,
        leading: (flags & 2) !== 0,
        trailing: (flags & 4) !== 0,
    };
}
export function whitespaceFrame(fields: readonly string[]): WhitespaceFrameInterface {
    const samples: CJSpaceSampleInterface[] = [];
    for(const item of (fields[21] ?? '').split(';')) {
        if(item === '') continue;
        const values = item.split(',');
        samples.push({
            type: values[0] ?? '',
            value: number(values[1] ?? ''),
            previousKind: values[2] ?? '',
            nextKind: values[3] ?? '',
        });
    }
    const kinds: string[] = [];
    const setext: boolean[] = [];
    if((fields[19] ?? '') !== '') {
        for(const kind of (fields[19] ?? '').split(',')) kinds.push(kind);
        for(const flag of (fields[20] ?? '').split(',')) setext.push(flag === '1');
    }
    return {
        value: decode(fields[1] ?? ''),
        proseWrap: fields[2] ?? '',
        link: fields[3] === '1',
        previous: token(fields, 4),
        next: token(fields, 9),
        afterNext: token(fields, 14),
        ancestorKinds: kinds,
        ancestorSetext: setext,
        samples,
    };
}
