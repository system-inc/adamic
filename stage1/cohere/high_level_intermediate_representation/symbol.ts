// Resident compiler facts needed by lower.go:1259; no lexical binding surrogate.
import { panic } from 'adamic';
import type { Checker } from '../lint/checker.a';
import { Frames, header } from '../typeaware/frames.ts';

export type SymbolDeclarationType = {
    readonly kind: string;
    readonly sameSource: boolean;
    readonly start: number;
    readonly end: number;
    readonly name: string;
    readonly property: string;
    readonly module: string;
};
export class SymbolFacts {
    readonly identity: number;
    readonly name: string;
    readonly declarations: readonly SymbolDeclarationType[];
    constructor(wire: string) {
        const frames = new Frames(wire);
        header(frames, 'symbol');
        this.identity = frames.natural();
        this.name = frames.field();
        const count = frames.natural();
        const declarations: SymbolDeclarationType[] = [];
        for(let index = 0; index < count; index++) {
            declarations.push({ kind: frames.field(), sameSource: frames.yes(), start: frames.natural(), end: frames.natural(), name: frames.field(), property: frames.field(), module: frames.field() });
        }
        frames.end();
        if(this.identity === 0 && (this.name !== '' || count !== 0)) { panic('absent symbol has declarations'); }
        this.declarations = declarations;
    }
}
export function readSymbol(checker: Checker, node: number): SymbolFacts {
    const answer = checker.ask(node, 'symbol');
    return new SymbolFacts(answer.value ?? panic(`HIR symbol query refused: ${answer.reason}`));
}
