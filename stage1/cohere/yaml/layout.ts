// Go cohere's reachable YAML document layout, with identity preserved by arena indices.
import { panic } from 'adamic';
import { LayoutNode } from './layoutNode.ts';
import { LayoutCommand } from './layoutCommand.ts';
import { LayoutIndent } from './layoutIndent.ts';
import { stringWidth } from './width.ts';
export class Layout {
    nodes: LayoutNode[] = [];
    indents: LayoutIndent[] = [new LayoutIndent()];
    modes: number[] = [];
    output: string[] = [];
    column = 0;
    remeasure = false;
    get(index: number): LayoutNode {
        return this.nodes[index] ?? panic('missing layout node');
    }
    indentAt(index: number): LayoutIndent {
        return this.indents[index] ?? panic('missing layout indentation');
    }
    make(kind: string, parts: number[]): number {
        const index = this.nodes.length;
        const node = new LayoutNode(kind);
        node.parts = parts;
        this.nodes.push(node);
        return index;
    }
    text(text: string): number {
        const result = this.make('text', []);
        this.get(result).text = text;
        this.get(result).width = stringWidth(text);
        return result;
    }
    concat(parts: number[]): number {
        return this.make('concat', parts);
    }
    join(separator: number, parts: readonly number[]): number {
        const result: number[] = [];
        for(let index = 0; index < parts.length; index++) {
            if(index > 0) result.push(separator);
            result.push(parts[index] ?? -1);
        }
        return this.concat(result);
    }
    line(soft: boolean, hard: boolean, literal: boolean): number {
        const result = this.make('line', []);
        this.get(result).soft = soft;
        this.get(result).broken = hard;
        this.get(result).literal = literal;
        return hard ? this.concat([result, this.make('break', [])]) : result;
    }
    group(content: number, id: number, conditional: boolean): number {
        const result = this.make('group', [content]);
        this.get(result).group = id;
        this.get(result).conditional = conditional;
        return result;
    }
    ifBreak(broken: number, flat: number, group: number): number {
        const result = this.make('ifBreak', [broken, flat]);
        this.get(result).group = group;
        return result;
    }
    align(amount: number, content: number): number {
        const result = this.make('align', [content]);
        this.get(result).amount = amount;
        return result;
    }
    /** @mutates groups The shared traversal stack tracks enclosing groups during break propagation. */
    propagate(index: number, groups: number[]): void {
        const node = this.get(index);
        if(node.kind === 'break' && groups.length > 0) {
            const parent = this.get(groups[groups.length - 1] ?? -1);
            if(!parent.conditional) parent.broken = true;
        }
        if(node.kind === 'group') groups.push(index);
        for(const part of node.parts) this.propagate(part, groups);
        if(node.kind === 'group') {
            groups.pop();
            if(node.broken && groups.length > 0) {
                const parent = this.get(groups[groups.length - 1] ?? -1);
                if(!parent.conditional) parent.broken = true;
            }
        }
    }
    trailing(text: string, previous: number): number {
        let count = 0;
        for(let index = text.length - 1; index >= 0; index--) {
            const code = text.charCodeAt(index);
            if(code !== 32 && code !== 9) return count;
            count++;
        }
        return previous + count;
    }
    fits(
        next: LayoutCommand,
        rest: readonly LayoutCommand[],
        width: number,
        suffix: boolean,
        mustFlat: boolean,
    ): boolean {
        const stack: LayoutCommand[] = [next];
        let restIndex = rest.length;
        let remaining = width;
        let pending = false;
        let trailing = 0;
        let hasSuffix = suffix;
        while(remaining >= 0) {
            if(stack.length === 0) {
                if(restIndex === 0) return true;
                restIndex--;
                stack.push(rest[restIndex] ?? panic('missing fit rest'));
                continue;
            }
            const current = stack.pop() ?? panic('missing fit command');
            const node = this.get(current.document);
            if(node.kind === 'text') {
                if(node.text !== '') {
                    if(pending) {
                        trailing++;
                        remaining--;
                        pending = false;
                    }
                    trailing = this.trailing(node.text, trailing);
                    remaining -= node.width;
                }
            }
            else if(node.kind === 'group') {
                if(mustFlat && node.broken) return false;
                stack.push(new LayoutCommand(node.parts[0] ?? -1, 0, node.broken ? 1 : current.mode));
            }
            else if(node.kind === 'ifBreak') {
                let mode = current.mode;
                if(node.group >= 0) {
                    mode = this.modes[node.group] ?? 0;
                    if(mode === 0) mode = 2;
                }
                stack.push(new LayoutCommand(node.parts[mode === 1 ? 0 : 1] ?? -1, 0, current.mode));
            }
            else if(node.kind === 'line') {
                if(current.mode === 1 || node.broken) return true;
                if(!node.soft) pending = true;
            }
            else if(node.kind === 'suffix') hasSuffix = true;
            else if(node.kind === 'boundary' && hasSuffix) return false;
            else if(node.kind !== 'break') {
                for(let index = node.parts.length - 1; index >= current.offset; index--)
                    stack.push(new LayoutCommand(node.parts[index] ?? -1, 0, current.mode));
            }
        }
        return false;
    }
    aligned(base: number, amount: number): number {
        const original = this.indentAt(base);
        if(amount === 0) return base;
        if(amount === -999) return original.root;
        const result = new LayoutIndent();
        result.queue = original.queue.slice();
        result.root = original.root;
        if(amount === -998) result.root = base;
        else if(amount < 0) result.queue.pop();
        else result.queue.push(amount);
        for(const width of result.queue) result.width += width;
        this.indents.push(result);
        return this.indents.length - 1;
    }
    trim(): number {
        let count = 0;
        while(this.output.length > 0) {
            const last = this.output.pop() ?? '';
            let end = last.length;
            while(end > 0 && (last.charCodeAt(end - 1) === 32 || last.charCodeAt(end - 1) === 9)) end--;
            count += last.length - end;
            if(end > 0) {
                this.output.push(last.slice(0, end));
                break;
            }
        }
        return count;
    }
    print(root: number): string {
        this.propagate(root, []);
        const stack: LayoutCommand[] = [new LayoutCommand(root, 0, 1)];
        let suffixes: LayoutCommand[] = [];
        while(stack.length > 0) {
            const current = stack.pop() ?? panic('missing layout command');
            const node = this.get(current.document);
            if(node.kind === 'text') {
                if(node.text !== '') {
                    this.output.push(node.text);
                    if(stack.length > 0) this.column += node.width;
                }
            }
            else if(node.kind === 'group') {
                const content = node.parts[0] ?? -1;
                let mode = node.broken ? 1 : 2;
                if(current.mode !== 2 || this.remeasure) {
                    this.remeasure = false;
                    if(
                        node.broken ||
                        !this.fits(
                            new LayoutCommand(content, current.indentation, 2),
                            stack,
                            80 - this.column,
                            suffixes.length > 0,
                            false,
                        )
                    )
                        mode = 1;
                }
                if(node.group >= 0) {
                    while(this.modes.length <= node.group) this.modes.push(0);
                    this.modes[node.group] = mode;
                }
                stack.push(new LayoutCommand(content, current.indentation, mode));
            }
            else if(node.kind === 'align')
                stack.push(
                    new LayoutCommand(
                        node.parts[0] ?? -1,
                        this.aligned(current.indentation, node.amount),
                        current.mode,
                    ),
                );
            else if(node.kind === 'ifBreak') {
                const mode = node.group >= 0 ? (this.modes[node.group] ?? 0) : current.mode;
                if(mode !== 0)
                    stack.push(
                        new LayoutCommand(node.parts[mode === 1 ? 0 : 1] ?? -1, current.indentation, current.mode),
                    );
            }
            else if(node.kind === 'suffix')
                suffixes.push(new LayoutCommand(node.parts[0] ?? -1, current.indentation, current.mode));
            else if(node.kind === 'line') {
                if(current.mode === 2 && !node.broken) {
                    if(!node.soft) {
                        this.output.push(' ');
                        this.column++;
                    }
                }
                else if(suffixes.length > 0) {
                    stack.push(current);
                    for(let index = suffixes.length - 1; index >= 0; index--)
                        stack.push(suffixes[index] ?? panic('missing suffix'));
                    suffixes = [];
                }
                else {
                    if(current.mode === 2) this.remeasure = true;
                    const indentation = this.indentAt(current.indentation);
                    if(!node.literal) this.trim();
                    const width = node.literal ? this.indentAt(indentation.root).width : indentation.width;
                    this.output.push('\n');
                    this.output.push(' '.repeat(width));
                    this.column = width;
                }
            }
            else if(node.kind === 'fill') {
                const remaining = node.parts.length - current.offset;
                if(remaining > 0) {
                    const first = node.parts[current.offset] ?? -1;
                    const flat = this.fits(
                        new LayoutCommand(first, current.indentation, 2),
                        [],
                        80 - this.column,
                        suffixes.length > 0,
                        true,
                    );
                    let spaceFlat = flat;
                    if(remaining > 2) {
                        const combination = this.concat([
                            first,
                            node.parts[current.offset + 1] ?? -1,
                            node.parts[current.offset + 2] ?? -1,
                        ]);
                        spaceFlat = this.fits(
                            new LayoutCommand(combination, current.indentation, 2),
                            [],
                            80 - this.column,
                            suffixes.length > 0,
                            true,
                        );
                        const next = new LayoutCommand(current.document, current.indentation, current.mode);
                        next.offset = current.offset + 2;
                        stack.push(next);
                    }
                    if(remaining > 1)
                        stack.push(
                            new LayoutCommand(
                                node.parts[current.offset + 1] ?? -1,
                                current.indentation,
                                spaceFlat ? 2 : 1,
                            ),
                        );
                    stack.push(new LayoutCommand(first, current.indentation, spaceFlat || flat ? 2 : 1));
                }
            }
            else if(node.kind !== 'break') {
                for(let index = node.parts.length - 1; index >= 0; index--)
                    stack.push(new LayoutCommand(node.parts[index] ?? -1, current.indentation, current.mode));
            }
            if(stack.length === 0 && suffixes.length > 0) {
                for(let index = suffixes.length - 1; index >= 0; index--)
                    stack.push(suffixes[index] ?? panic('missing final suffix'));
                suffixes = [];
            }
        }
        return this.output.join('');
    }
}
