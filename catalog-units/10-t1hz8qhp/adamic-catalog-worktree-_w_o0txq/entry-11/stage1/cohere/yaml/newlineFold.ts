import { char } from './scalarText.ts';
export class NewlineFold {
    value = '';
    offset: number;
    constructor(source: string, start: number) {
        let offset = start;
        let next = char(source, offset + 1);
        while(next === ' ' || next === '\t' || next === '\n' || next === '\r') {
            if(next === '\r' && char(source, offset + 2) !== '\n') break;
            if(next === '\n') this.value += '\n';
            offset++;
            next = char(source, offset + 1);
        }
        if(this.value === '') this.value = ' ';
        this.offset = offset;
    }
}
