import * as ts from "./typescript.js";
export function probe(file: ts.SourceFile): ts.LineAndCharacter {
    return ts.getLineAndCharacterOfPosition(file, 0);
}
