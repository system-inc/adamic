import type { Rules } from './rules.ts';
import type { Types } from './types.ts';
import type { TypeFact } from './type_fact.ts';
import { types } from './facts.ts';
import { Frames, header } from './frames.ts';

export class TypeMetadata {
    readonly objectFlags: number;
    readonly symbol: number;
    readonly targetFlags: number;
    readonly targetSymbol: number;
    readonly readonlyTuple: boolean;
    constructor(flags: number, symbol: number, targetFlags: number, targetSymbol: number, readonlyTuple: boolean) {
        this.objectFlags = flags;
        this.symbol = symbol;
        this.targetFlags = targetFlags;
        this.targetSymbol = targetSymbol;
        this.readonlyTuple = readonlyTuple;
    }
    isClass(): boolean {
        return (this.objectFlags & 1) !== 0 || ((this.objectFlags & 4) !== 0 && (this.targetFlags & 1) !== 0);
    }
    classSymbol(): number {
        return (this.objectFlags & 4) !== 0 ? this.targetSymbol : this.symbol;
    }
}
export class Properties {
    readonly names: readonly string[];
    readonly flags: readonly number[];
    readonly keys: readonly number[];
    readonly readonlys: readonly boolean[];
    readonly types: Types;
    constructor(names: readonly string[], keys: readonly number[], flags: readonly number[], readonlys: readonly boolean[], types_: Types) {
        this.names = names;
        this.keys = keys;
        this.flags = flags;
        this.readonlys = readonlys;
        this.types = types_;
    }
}
export class CheckerFacts {
    readonly rules: Rules;
    constructor(rules: Rules) { this.rules = rules; }
    metadata(index: number, type: TypeFact): TypeMetadata {
        const frames = new Frames(this.rules.ask(index, `type-metadata\n${type.id}`));
        header(frames, 'type-metadata');
        const result = new TypeMetadata(frames.natural(), frames.natural(), frames.natural(), frames.natural(), frames.yes());
        frames.end();
        return result;
    }
    properties(index: number, type: TypeFact): Properties {
        const frames = new Frames(this.rules.ask(index, `type-properties\n${type.id}`));
        header(frames, 'type-properties');
        const count = frames.natural();
        const names: string[] = [];
        const flags: number[] = [];
        const keys: number[] = [];
        const readonlys: boolean[] = [];
        for(let at = 0; at < count; at++) {
            names.push(frames.field()); keys.push(frames.natural()); flags.push(frames.natural()); readonlys.push(frames.yes());
        }
        const graph = types(`1\n115\ntype-properties${frames.text.slice(frames.cursor)}`, 'type-properties');
        return new Properties(names, keys, flags, readonlys, graph);
    }
    compare(index: number, left: number, right: number, identical = false): boolean {
        if(left === right) { return true; }
        const question = identical ? 'identical-types' : 'assignable-types';
        const frames = new Frames(this.rules.ask(index, `${question}\n${left}\n${right}`));
        header(frames, question);
        const result = frames.yes(); frames.end(); return result;
    }
}
