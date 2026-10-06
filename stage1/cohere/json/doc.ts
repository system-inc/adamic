// JSON's reachable subset of cohere/internal/format/doc: concat, indent, group, lines and fill.
// Documents refer to children by table index, keeping the ownership graph acyclic.
import { panic } from 'adamic';
import { stringWidth } from './width.ts';
export interface DocInterface {
    kind: string;
    text: string;
    parts: number[];
    broken: boolean;
    width: number;
}
interface CommandInterface {
    doc: number;
    indent: number;
    flat: boolean;
    offset: number;
}
export class Documents {
    readonly nodes: DocInterface[] = [];
    add(kind: string, text: string, parts: number[], broken: boolean): number {
        let mustBreak = broken;
        if(kind !== 'line') {
            for(const part of parts) {
                if(this.get(part).broken) {
                    mustBreak = true;
                }
            }
        }
        const index = this.nodes.length;
        this.nodes.push({ kind, text, parts, broken: mustBreak, width: stringWidth(text) });
        return index;
    }
    text(text: string): number {
        return this.add('text', text, [], false);
    }
    concat(parts: number[]): number {
        return this.add('concat', '', parts, false);
    }
    line(soft: boolean, hard: boolean): number {
        return this.add('line', soft ? '' : ' ', [], hard);
    }
    indent(child: number): number {
        return this.add('indent', '', [child], false);
    }
    group(child: number, broken: boolean): number {
        return this.add('group', '', [child], broken);
    }
    fill(parts: number[]): number {
        return this.add('fill', '', parts, false);
    }
    get(index: number): DocInterface {
        return this.nodes[index] ?? panic(`missing document ${index}`);
    }
    // fits examines a candidate and the pending command stack up to the next real line.
    fits(
        candidate: readonly CommandInterface[],
        rest: readonly CommandInterface[],
        width: number,
        mustFlat: boolean,
    ): boolean {
        let remainingWidth = width;
        const stack: CommandInterface[] = [...candidate];
        let restIndex = rest.length;
        while(remainingWidth >= 0) {
            if(stack.length === 0) {
                if(restIndex === 0) {
                    return true;
                }
                restIndex--;
                stack.push(rest[restIndex] ?? panic('missing rest command'));
                continue;
            }
            const command = stack.pop() ?? panic('missing fit command');
            const node = this.get(command.doc);
            if(node.kind === 'suffix') {
                continue;
            }
            if(node.kind === 'text') {
                remainingWidth -= node.width;
            }
            else if(node.kind === 'line') {
                if(!command.flat || node.broken) {
                    return true;
                }
                remainingWidth -= node.width;
            }
            else {
                if(node.kind === 'group' && mustFlat && node.broken) {
                    return false;
                }
                const flat = node.kind === 'group' && node.broken ? false : command.flat;
                for(let index = node.parts.length - 1; index >= command.offset; index--) {
                    stack.push({ doc: node.parts[index] ?? panic('missing fit part'), indent: 0, flat, offset: 0 });
                }
            }
        }
        return false;
    }
    print(root: number): string {
        const stack: CommandInterface[] = [{ doc: root, indent: 0, flat: false, offset: 0 }];
        const output: string[] = [];
        let column = 0;
        let pendingIndent = 0;
        const indentation: string[] = [''];
        const suffixes: CommandInterface[] = [];
        while(stack.length > 0) {
            const command = stack.pop() ?? panic('missing print command');
            const node = this.get(command.doc);
            if(node.kind === 'text') {
                if(node.text !== '' && pendingIndent > 0) {
                    while(indentation.length <= pendingIndent) {
                        indentation.push(`${indentation[indentation.length - 1] ?? panic('missing indentation')} `);
                    }
                    output.push(indentation[pendingIndent] ?? panic('missing indentation'));
                    pendingIndent = 0;
                }
                output.push(node.text);
                column += node.width;
            }
            else if(node.kind === 'suffix') {
                suffixes.push({ doc: this.concat(node.parts), indent: command.indent, flat: command.flat, offset: 0 });
            }
            else if(node.kind === 'line') {
                if(suffixes.length > 0 && (!command.flat || node.broken)) {
                    stack.push(command);
                    while(suffixes.length > 0) {
                        stack.push(suffixes.pop() ?? panic('missing suffix'));
                    }
                    continue;
                }
                if(command.flat && !node.broken) {
                    output.push(node.text);
                    column += node.width;
                }
                else {
                    output.push('\n');
                    pendingIndent = command.indent;
                    column = command.indent;
                }
            }
            else if(node.kind === 'group') {
                const child = node.parts[0] ?? panic('missing group child');
                const candidate: CommandInterface = { doc: child, indent: command.indent, flat: true, offset: 0 };
                const flat = !node.broken && (command.flat || this.fits([candidate], stack, 80 - column, false));
                stack.push({ doc: child, indent: command.indent, flat, offset: 0 });
            }
            else if(node.kind === 'fill') {
                const index = command.offset;
                const first = node.parts[index] ?? panic('missing fill part');
                let flat = this.fits([{ doc: first, indent: 0, flat: true, offset: 0 }], [], 80 - column, true);
                if(index + 1 < node.parts.length) {
                    const whitespace = node.parts[index + 1] ?? panic('missing fill space');
                    let spaceFlat = flat;
                    if(index + 2 < node.parts.length) {
                        const second = node.parts[index + 2] ?? panic('missing next fill part');
                        spaceFlat = this.fits(
                            [
                                { doc: second, indent: 0, flat: true, offset: 0 },
                                { doc: whitespace, indent: 0, flat: true, offset: 0 },
                                { doc: first, indent: 0, flat: true, offset: 0 },
                            ],
                            [],
                            80 - column,
                            true,
                        );
                        if(spaceFlat) {
                            flat = true;
                        }
                        stack.push({ doc: command.doc, indent: command.indent, flat: command.flat, offset: index + 2 });
                    }
                    stack.push({ doc: whitespace, indent: command.indent, flat: spaceFlat, offset: 0 });
                }
                stack.push({ doc: first, indent: command.indent, flat, offset: 0 });
            }
            else {
                const indent = node.kind === 'indent' ? command.indent + 2 : command.indent;
                for(let index = node.parts.length - 1; index >= 0; index--) {
                    stack.push({
                        doc: node.parts[index] ?? panic('missing concat part'),
                        indent,
                        flat: command.flat,
                        offset: 0,
                    });
                }
            }
        }
        return output.join('');
    }
}
