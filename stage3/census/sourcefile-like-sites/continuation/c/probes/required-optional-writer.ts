export interface LineMapSource { readonly text: string; lineMap?: readonly number[] | undefined; }
export interface SourceFile { readonly text: string; lineMap: readonly number[] | undefined; readonly kind: 307; }
function computeLineStarts(text: string): readonly number[] { return [0, text.length]; }
export function getLineStarts(sourceFile: LineMapSource): readonly number[] {
    return sourceFile.lineMap || (sourceFile.lineMap = computeLineStarts(sourceFile.text));
}
const file: SourceFile = { text: "abc", lineMap: undefined, kind: 307 };
export const result = getLineStarts(file);
