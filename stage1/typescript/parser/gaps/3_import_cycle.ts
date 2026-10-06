import { value } from './3_import_cycle_peer.ts';
export class Entry {
    readonly number = 1;
}
console.log(`${value(new Entry())}`);
