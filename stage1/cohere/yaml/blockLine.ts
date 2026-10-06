import { char } from './scalarText.ts';
export class BlockLine {
    indent: string;
    content: string;
    constructor(source: string) {
        let index = 0;
        while(char(source, index) === ' ') index++;
        this.indent = source.slice(0, index);
        this.content = source.slice(index);
    }
}
