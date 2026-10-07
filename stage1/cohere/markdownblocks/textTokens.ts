// Owned splitText tokens; callers receive the completed readonly array.
export interface TextTokenInterface {
    readonly type: string;
    readonly value: string;
    readonly kind: string;
    readonly cj: boolean;
    readonly leading: boolean;
    readonly trailing: boolean;
}
export class TextTokens {
    readonly values: TextTokenInterface[] = [];
    whitespace(value: string): void {
        this.values.push({ type: 'whitespace', value, kind: '', cj: false, leading: false, trailing: false });
    }
    word(value: string, kind: string, cj: boolean, leading: boolean, trailing: boolean): void {
        const last = this.values[this.values.length - 1];
        if(last !== undefined && last.type === 'word') {
            const between =
                (last.kind === 'non-cjk' && kind === 'cjk-punctuation') ||
                (last.kind === 'cjk-punctuation' && kind === 'non-cjk');
            if(!between && !last.value.includes('\u3000') && !value.includes('\u3000')) this.whitespace('');
        }
        this.values.push({ type: 'word', value, kind, cj, leading, trailing });
    }
}
