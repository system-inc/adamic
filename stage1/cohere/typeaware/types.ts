import { panic } from 'adamic';
import { unionFlag } from './flags.ts';
import type { TypeFact } from './type_fact.ts';

export class Types {
    readonly strict: boolean;
    readonly present: boolean;
    readonly roots: readonly number[];
    readonly rests: readonly boolean[];
    readonly records: readonly TypeFact[];
    constructor(
        strict: boolean,
        present: boolean,
        roots: readonly number[],
        rests: readonly boolean[],
        records: readonly TypeFact[],
    ) {
        this.strict = strict;
        this.present = present;
        this.roots = roots;
        this.rests = rests;
        this.records = records;
    }
    type(id: number): TypeFact {
        for(const record of this.records) {
            if(record.id === id) {
                return record;
            }
        }
        panic('missing checker type identity');
    }
    root(index = 0): TypeFact {
        return this.type(this.roots[index] ?? panic('missing checker root'));
    }
    parts(type: TypeFact, mask = unionFlag): readonly number[] {
        return (type.flags & mask) !== 0 ? type.parts : [type.id];
    }
    has(type: TypeFact, mask: number): boolean {
        return this.parts(type).some((part) => (this.type(part).flags & mask) !== 0);
    }
}
