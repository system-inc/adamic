// Owned command stack for the Markdown document printer.
import { panic } from 'adamic';
export interface CommandInterface {
    readonly indent: number;
    readonly mode: number;
    readonly doc: number;
    readonly offset: number;
}
export function command(indentID: number, mode: number, doc: number, offset = 0): CommandInterface {
    return { indent: indentID, mode, doc, offset };
}
export class CommandStack {
    readonly values: CommandInterface[] = [];
    appendParts(parts: readonly number[], current: CommandInterface, offset = 0): void {
        for(let index = parts.length - 1; index >= offset; index--) {
            this.values.push(command(current.indent, current.mode, parts[index] ?? panic('document child'), 0));
        }
    }
}
