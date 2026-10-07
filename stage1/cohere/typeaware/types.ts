import { panic } from 'adamic';
import { unionFlag } from './flags.ts';
import type { TypeFact } from './type_fact.ts';

export class Types {
    readonly strict: boolean;
    readonly present: boolean;
    readonly roots: readonly number[];
    readonly rests: readonly boolean[];
    readonly records: readonly TypeFact[];
    readonly ordered: readonly TypeFact[];
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
        // Checker graphs are complete and read-only after decoding. Keep wire
        // order intact; the stable sorted copy preserves the first equal ID.
        const ordered = records.slice();
        ordered.sort((left, right) => left.id - right.id);
        this.ordered = ordered;
    }
    type(id: number): TypeFact {
        let first = 0;
        let last = this.ordered.length;
        while(first < last) {
            const middle = Math.floor((first + last) / 2);
            const record = this.ordered[middle] ?? panic('missing checker type index');
            if(record.id < id) { first = middle + 1; }
            else { last = middle; }
        }
        const record = this.ordered[first];
        if(record !== undefined && record.id === id) { return record; }
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
