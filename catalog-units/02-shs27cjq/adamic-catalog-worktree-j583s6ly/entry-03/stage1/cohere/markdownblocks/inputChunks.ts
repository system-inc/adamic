// micromark/lib/preprocess.js: one complete UTF-16 input, end=true.
export interface InputChunkInterface {
    text: readonly number[];
    code: number;
    isText: boolean;
}
function codeChunk(code: number): InputChunkInterface {
    return { text: [], code, isText: false };
}
function textChunk(text: readonly number[]): InputChunkInterface {
    return { text, code: 0, isText: true };
}
export function inputChunks(units: readonly number[]): InputChunkInterface[] {
    const chunks: InputChunkInterface[] = [];
    let column = 1;
    let carriage = false;
    let start = units[0] === 65279 ? 1 : 0;
    let buffer: readonly number[] = [];
    while(start < units.length) {
        let end = start;
        while(end < units.length) {
            const unit = units[end];
            if(unit === 0 || unit === 9 || unit === 10 || unit === 13) break;
            end++;
        }
        if(end === units.length) {
            buffer = units.slice(start);
            break;
        }
        const code = units[end];
        if(code === 10 && start === end && carriage) {
            chunks.push(codeChunk(-3));
            carriage = false;
        }
        else {
            if(carriage) {
                chunks.push(codeChunk(-5));
                carriage = false;
            }
            if(start < end) {
                chunks.push(textChunk(units.slice(start, end)));
                column += end - start;
            }
            if(code === 0) {
                chunks.push(codeChunk(65533));
                column++;
            }
            else if(code === 9) {
                const next = Math.ceil(column / 4) * 4;
                chunks.push(codeChunk(-2));
                while(column < next) {
                    chunks.push(codeChunk(-1));
                    column++;
                }
                column++;
            }
            else if(code === 10) {
                chunks.push(codeChunk(-4));
                column = 1;
            }
            else {
                carriage = true;
                column = 1;
            }
        }
        start = end + 1;
    }
    if(carriage) chunks.push(codeChunk(-5));
    if(buffer.length > 0) chunks.push(textChunk(buffer));
    chunks.push(codeChunk(0));
    return chunks;
}
export function serializeChunks(chunks: readonly InputChunkInterface[]): string {
    const output: string[] = [];
    for(const chunk of chunks) output.push(chunk.isText ? `T${chunk.text.join(',')}` : `C${chunk.code}`);
    return output.join(';');
}
