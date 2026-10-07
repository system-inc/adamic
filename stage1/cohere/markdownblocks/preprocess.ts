// Cohere micromark/parse.go preprocessing for one complete UTF-16 source.
export interface ChunkInterface {
    readonly isText: boolean;
    readonly text: string;
    readonly code: number;
}
export function preprocess(value: string): ChunkInterface[] {
    const chunks: ChunkInterface[] = [];
    let column = 1;
    let carriageReturn = false;
    let start = value.charCodeAt(0) === 0xfeff ? 1 : 0;
    let buffer = '';
    while(start < value.length) {
        let end = start;
        while(end < value.length) {
            const unit = value.charCodeAt(end);
            if(unit === 0 || unit === 9 || unit === 10 || unit === 13) break;
            end++;
        }
        if(end === value.length) {
            buffer = value.slice(start);
            break;
        }
        const code = value.charCodeAt(end);
        if(code === 10 && start === end && carriageReturn) {
            chunks.push({ isText: false, text: '', code: -3 });
            carriageReturn = false;
        }
        else {
            if(carriageReturn) {
                chunks.push({ isText: false, text: '', code: -5 });
                carriageReturn = false;
            }
            if(start < end) {
                chunks.push({ isText: true, text: value.slice(start, end), code: 0 });
                column += end - start;
            }
            if(code === 0) {
                chunks.push({ isText: false, text: '', code: 65533 });
                column++;
            }
            else if(code === 9) {
                const next = Math.floor((column + 3) / 4) * 4;
                chunks.push({ isText: false, text: '', code: -2 });
                while(column < next) {
                    chunks.push({ isText: false, text: '', code: -1 });
                    column++;
                }
                column++;
            }
            else if(code === 10) {
                chunks.push({ isText: false, text: '', code: -4 });
                column = 1;
            }
            else {
                carriageReturn = true;
                column = 1;
            }
        }
        start = end + 1;
    }
    if(carriageReturn) chunks.push({ isText: false, text: '', code: -5 });
    if(buffer !== '') chunks.push({ isText: true, text: buffer, code: 0 });
    chunks.push({ isText: false, text: '', code: 0 });
    return chunks;
}
