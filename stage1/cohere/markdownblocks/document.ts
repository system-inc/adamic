// Markdown document layout from cohere/internal/format/doc/printer.go.
// Nodes, indentation roots and commands use arena IDs, never owning back pointers.
// Text display widths are computed natively from cohere's pinned Unicode data.
import { panic } from 'adamic';
import { stringWidth } from './width.ts';
import { CommandStack, command, type CommandInterface } from './commandStack.ts';
export interface DocumentInterface {
    readonly kind: string;
    readonly text: string;
    readonly width: number;
    readonly children: readonly number[];
    readonly flag: number;
    readonly group: number;
    breaks: boolean;
}
export class DocumentArena {
    readonly nodes: DocumentInterface[] = [];
    add(kind: string, text: string, width: number, children: readonly number[], flag = 0, group = -1): number {
        const id = this.nodes.length;
        this.nodes.push({
            kind,
            text,
            width: kind === 't' ? stringWidth(text) : width,
            children,
            flag,
            group,
            breaks: (flag & 1) !== 0,
        });
        return id;
    }
    node(id: number): DocumentInterface {
        return this.nodes[id] ?? panic('document ID outside arena');
    }
    text(value: string): number {
        return this.add('t', value, 0, []);
    }
    concat(parts: readonly number[]): number {
        return this.add('a', '', 0, parts);
    }
    align(prefix: string, contents: number): number {
        return this.add('s', prefix, 0, [contents]);
    }
    hardline(): number {
        return this.concat([this.add('h', '', 0, [], 2), this.add('p', '', 0, [])]);
    }
}
interface IndentCommandInterface {
    readonly kind: string;
    readonly text: string;
    readonly width: number;
}
interface IndentationInterface {
    readonly value: string;
    readonly length: number;
    readonly queue: readonly IndentCommandInterface[];
    readonly root: number;
}
function child(node: DocumentInterface, index = 0): number {
    return node.children[index] ?? panic('missing document child');
}
function trailing(text: string, previous: number): number {
    let count = 0;
    for(let index = text.length - 1; index >= 0; index--) {
        const code = text.charCodeAt(index);
        if(code !== 32 && code !== 9) return count;
        count++;
    }
    return previous + count;
}
function propagate(arena: DocumentArena, root: number): void {
    const stack: number[] = [root];
    const groups: number[] = [];
    const visited = new Map<number, boolean>();
    while(stack.length > 0) {
        const id = stack.pop() ?? panic('traverse stack');
        if(id < 0) {
            const leaving = arena.node(-id - 1);
            if(leaving.kind === 'g') {
                groups.pop();
                if(leaving.breaks && groups.length > 0) {
                    const parent = arena.node(groups[groups.length - 1] ?? panic('group stack'));
                    if((parent.flag & 2) === 0) parent.breaks = true;
                }
            }
            continue;
        }
        const node = arena.node(id);
        stack.push(-id - 1);
        if(node.kind === 'p' && groups.length > 0) {
            const parent = arena.node(groups[groups.length - 1] ?? panic('group stack'));
            if((parent.flag & 2) === 0) parent.breaks = true;
        }
        if(node.kind === 'g') {
            groups.push(id);
            if(visited.has(id)) continue;
            visited.set(id, true);
        }
        for(let index = node.children.length - 1; index >= 0; index--)
            stack.push(node.children[index] ?? panic('traverse child'));
    }
}
function fits(
    arena: DocumentArena,
    next: CommandInterface,
    rest: readonly CommandInterface[],
    initialWidth: number,
    modes: Map<number, number>,
    mustBeFlat: boolean,
): boolean {
    const stack = new CommandStack();
    stack.values.push(next);
    let remaining = initialWidth;
    let restIndex = rest.length;
    let pending = false;
    let tail = 0;
    while(remaining >= 0) {
        if(stack.values.length === 0) {
            if(restIndex === 0) return true;
            restIndex--;
            stack.values.push(rest[restIndex] ?? panic('fits rest'));
            continue;
        }
        const current = stack.values.pop() ?? panic('fits stack');
        const node = arena.node(current.doc);
        if(node.kind === 't') {
            if(node.text !== '') {
                if(pending) {
                    tail++;
                    remaining--;
                    pending = false;
                }
                tail = trailing(node.text, tail);
                remaining -= node.width;
            }
        }
        else if(node.kind === 'a' || node.kind === 'f') stack.appendParts(node.children, current, current.offset);
        else if(
            node.kind === 'i' ||
            node.kind === 's' ||
            node.kind === 'w' ||
            node.kind === 'r' ||
            node.kind === 'd' ||
            node.kind === 'l'
        )
            stack.values.push(command(current.indent, current.mode, child(node)));
        else if(node.kind === 'x') {
            remaining += tail;
            tail = 0;
        }
        else if(node.kind === 'g') {
            if(mustBeFlat && node.breaks) return false;
            const mode = node.breaks ? 1 : current.mode;
            stack.values.push(command(current.indent, mode, child(node, mode === 1 ? node.children.length - 1 : 0)));
        }
        else if(node.kind === 'b') {
            const mode = node.group < 0 ? current.mode : (modes.get(node.group) ?? 2);
            stack.values.push(command(current.indent, current.mode, child(node, mode === 1 ? 0 : 1)));
        }
        else if(node.kind === 'h') {
            if(current.mode === 1 || (node.flag & 2) !== 0) return true;
            if((node.flag & 1) === 0) pending = true;
        }
        else if(node.kind !== 'p') panic('unsupported Markdown document in fits');
    }
    return false;
}
function indent(
    arena: readonly IndentationInterface[],
    baseID: number,
    kind: string,
    text: string,
    width: number,
    tabWidth: number,
    useTabs: boolean,
): IndentationInterface {
    const base = arena[baseID] ?? panic('indentation ID');
    if(kind === 'r') {
        return { value: base.value, length: base.length, queue: base.queue, root: baseID };
    }
    if(kind === 'd') return arena[base.root] ?? panic('indent root');
    if((kind === 's' && text === '') || (kind === 'w' && width === 0)) return base;
    const queue =
        kind === 'w' && width < 0 ? base.queue.slice(0, Math.max(0, base.queue.length - 1)) : base.queue.slice();
    if(!(kind === 'w' && width < 0)) queue.push({ kind, text, width });
    const parts: string[] = [];
    let length = 0;
    let lastTabs = 0;
    let lastSpaces = 0;
    for(const queued of queue) {
        if(queued.kind === 'w') {
            lastTabs++;
            lastSpaces += queued.width;
            continue;
        }
        parts.push(useTabs ? '\t'.repeat(lastTabs) : ' '.repeat(lastSpaces));
        length += useTabs ? lastTabs * tabWidth : lastSpaces;
        lastTabs = 0;
        lastSpaces = 0;
        if(queued.kind === 'i') {
            parts.push(useTabs ? '\t' : ' '.repeat(tabWidth));
            length += tabWidth;
        }
        else if(queued.kind === 's') {
            parts.push(queued.text);
            length += queued.text.length;
        }
        else panic('indent command');
    }
    parts.push(' '.repeat(lastSpaces));
    length += lastSpaces;
    return { value: parts.join(''), length, queue, root: base.root };
}
export function printDocument(
    arena: DocumentArena,
    root: number,
    printWidth: number,
    tabWidth: number,
    useTabs = false,
): string {
    propagate(arena, root);
    const indents: IndentationInterface[] = [{ value: '', length: 0, queue: [], root: 0 }];
    const stack = new CommandStack();
    stack.values.push(command(0, 1, root));
    const modes = new Map<number, number>();
    const output: string[] = [];
    let position = 0;
    let remeasure = false;
    while(stack.values.length > 0) {
        const current = stack.values.pop() ?? panic('print stack');
        const node = arena.node(current.doc);
        if(node.kind === 't') {
            if(node.text !== '') {
                output.push(node.text);
                if(stack.values.length > 0) position += node.width;
            }
        }
        else if(node.kind === 'a') stack.appendParts(node.children, current);
        else if(node.kind === 'i' || node.kind === 's' || node.kind === 'w' || node.kind === 'r' || node.kind === 'd') {
            const aligned = indents.length;
            indents.push(indent(indents, current.indent, node.kind, node.text, node.width, tabWidth, useTabs));
            stack.values.push(command(aligned, current.mode, child(node)));
        }
        else if(node.kind === 'g') {
            let mode = 1;
            let contents = child(node);
            if(current.mode === 2 && !remeasure) mode = node.breaks ? 1 : 2;
            else {
                remeasure = false;
                if(
                    !node.breaks &&
                    fits(arena, command(current.indent, 2, contents), stack.values, printWidth - position, modes, false)
                )
                    mode = 2;
                else if((node.flag & 2) !== 0) {
                    contents = child(node, node.children.length - 1);
                    if(!node.breaks) {
                        for(let index = 1; index < node.children.length - 1; index++) {
                            const candidate = child(node, index);
                            if(
                                fits(
                                    arena,
                                    command(current.indent, 2, candidate),
                                    stack.values,
                                    printWidth - position,
                                    modes,
                                    false,
                                )
                            ) {
                                contents = candidate;
                                mode = 2;
                                break;
                            }
                        }
                    }
                }
            }
            stack.values.push(command(current.indent, mode, contents));
            if(node.group >= 0) modes.set(node.group, mode);
        }
        else if(node.kind === 'f') {
            const length = node.children.length - current.offset;
            if(length === 0) continue;
            const content = node.children[current.offset] ?? panic('fill content');
            const flat = command(current.indent, 2, content);
            const broken = command(current.indent, 1, content);
            const contentFits = fits(arena, flat, [], printWidth - position, modes, true);
            if(length === 1) {
                stack.values.push(contentFits ? flat : broken);
                continue;
            }
            const whitespace = node.children[current.offset + 1] ?? panic('fill whitespace');
            let together = false;
            if(length > 2) {
                const second = node.children[current.offset + 2] ?? panic('fill second');
                together = fits(
                    arena,
                    command(current.indent, 2, arena.concat([content, whitespace, second])),
                    [],
                    printWidth - position,
                    modes,
                    true,
                );
                stack.values.push(command(current.indent, current.mode, current.doc, current.offset + 2));
            }
            const flatSpace = length === 2 ? contentFits : together;
            stack.values.push(command(current.indent, flatSpace ? 2 : 1, whitespace));
            stack.values.push(together || contentFits ? flat : broken);
        }
        else if(node.kind === 'b') {
            const mode = node.group < 0 ? current.mode : (modes.get(node.group) ?? 0);
            if(mode === 1 || mode === 2)
                stack.values.push(command(current.indent, current.mode, child(node, mode === 1 ? 0 : 1)));
        }
        else if(node.kind === 'l') stack.values.push(command(current.indent, current.mode, child(node)));
        else if(node.kind === 'h' || node.kind === 'x') {
            if(node.kind === 'h' && current.mode === 2 && (node.flag & 2) === 0) {
                if((node.flag & 1) === 0) {
                    output.push(' ');
                    position++;
                }
                continue;
            }
            if(node.kind === 'h' && current.mode === 2) remeasure = true;
            if(node.kind === 'h' && (node.flag & 4) !== 0) {
                const base = indents[current.indent] ?? panic('literal indent');
                const rootIndent = indents[base.root] ?? panic('literal root');
                output.push('\n');
                output.push(rootIndent.value);
                position = rootIndent.length;
            }
            else {
                // Trim trailing spaces/tabs across text fragments without joining the whole buffer.
                let removed = 0;
                while(output.length > 0) {
                    const last = output.pop() ?? panic('print result');
                    let end = last.length;
                    while(end > 0 && (last.charCodeAt(end - 1) === 32 || last.charCodeAt(end - 1) === 9)) end--;
                    removed += last.length - end;
                    if(end > 0) {
                        output.push(last.slice(0, end));
                        break;
                    }
                }
                position -= removed;
                if(node.kind === 'h') {
                    const base = indents[current.indent] ?? panic('line indent');
                    output.push('\n');
                    output.push(base.value);
                    position = base.length;
                }
            }
        }
        else if(node.kind !== 'p') panic('unsupported Markdown document in printer');
    }
    return output.join('');
}
