// cohere/internal/format/doc: indexed docs preserve shared-group identity without owning cycles.
import { panic } from 'adamic';
import { stringWidth } from './width.ts';
const supportedKinds: readonly string[] = [
    'text',
    'concat',
    'group',
    'conditionalGroup',
    'indent',
    'alignWidth',
    'alignString',
    'markAsRoot',
    'dedentToRoot',
    'trim',
    'fill',
    'ifBreak',
    'indentIfBreak',
    'lineSuffix',
    'lineSuffixBoundary',
    'line',
    'softline',
    'hardlineWithoutBreakParent',
    'literallineWithoutBreakParent',
    'label',
    'breakParent',
];
export interface SettingsOptions {
    readonly printWidth: number;
    readonly tabWidth: number;
    readonly useTabs: boolean;
}
export interface DocInterface {
    readonly kind: string;
    readonly text: string;
    readonly parts: readonly number[];
    readonly number: number;
    readonly key: string;
    broken: boolean;
}
interface IndentInterface {
    readonly value: string;
    readonly length: number;
    readonly queue: readonly number[];
    readonly strings: readonly string[];
    readonly root: number;
}
interface CommandInterface {
    readonly doc: number;
    readonly indent: number;
    readonly mode: number; // 1 broken, 2 flat; a missing keyed group's mode is 0.
    readonly offset: number;
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
export class Documents {
    readonly nodes: DocInterface[] = [];
    readonly settings: SettingsOptions;
    constructor(settings: SettingsOptions) {
        this.settings = settings;
    }
    get(index: number): DocInterface {
        return this.nodes[index] ?? panic(`missing doc ${index}`);
    }
    add(kind: string, parts: readonly number[], text = '', number = 0, key = '', broken = false): number {
        if(!supportedKinds.includes(kind)) panic(`unknown document kind ${kind}`);
        for(const child of parts)
            if(child < 0 || child >= this.nodes.length) panic('document edges must refer to earlier nodes');
        this.nodes.push({ kind, parts, text, number, key, broken });
        return this.nodes.length - 1;
    }
    text(value: string): number {
        return this.add('text', [], value);
    }
    concat(parts: readonly number[]): number {
        return this.add('concat', parts);
    }
    group(contents: number, key = '', broken = false): number {
        return this.add('group', [contents], '', 0, key, broken);
    }
    indent(contents: number): number {
        return this.add('indent', [contents]);
    }
    line(): number {
        return this.add('line', []);
    }
    softline(): number {
        return this.add('softline', []);
    }
    hardline(): number {
        return this.concat([this.add('hardlineWithoutBreakParent', []), this.add('breakParent', [])]);
    }
    ifBreak(broken: number, flat: number, key = ''): number {
        return this.add('ifBreak', [broken, flat], '', 0, key);
    }
    fill(parts: readonly number[]): number {
        return this.add('fill', parts);
    }
    join(separator: number, parts: readonly number[]): number {
        const joined: number[] = [];
        for(const part of parts) {
            if(joined.length > 0) joined.push(separator);
            joined.push(part);
        }
        return this.concat(joined);
    }
    textContent(index: number): string | undefined {
        const doc = this.get(index);
        if(doc.kind === 'text') return doc.text;
        if(doc.kind !== 'concat') return undefined;
        let text = '';
        for(const child of doc.parts) {
            const content = this.textContent(child);
            if(content === undefined) return undefined;
            text += content;
        }
        return text;
    }
    willBreak(index: number): boolean {
        const doc = this.get(index);
        if(
            (['group', 'conditionalGroup'].includes(doc.kind) && doc.broken) ||
            ['breakParent', 'hardlineWithoutBreakParent', 'literallineWithoutBreakParent'].includes(doc.kind)
        )
            return true;
        const count = doc.kind === 'conditionalGroup' ? Math.min(1, doc.parts.length) : doc.parts.length;
        for(let position = 0; position < count; position++)
            if(this.willBreak(doc.parts[position] ?? panic('missing break child'))) return true;
        return false;
    }
    removeLines(index: number): number {
        const doc = this.get(index);
        if(doc.kind === 'line') return this.text(' ');
        if(doc.kind === 'softline') return this.text('');
        if(doc.kind === 'ifBreak') return this.removeLines(doc.parts[1] ?? panic('missing flat branch'));
        const parts: number[] = [];
        for(const child of doc.parts) parts.push(this.removeLines(child));
        return this.add(doc.kind, parts, doc.text, doc.number, doc.key, doc.broken);
    }
    canBreak(index: number): boolean {
        const doc = this.get(index);
        if(['line', 'softline', 'hardlineWithoutBreakParent', 'literallineWithoutBreakParent'].includes(doc.kind))
            return true;
        for(const child of doc.parts) if(this.canBreak(child)) return true;
        return false;
    }
    propagate(root: number): void {
        const visited = new Set<number>();
        const groups: number[] = [];
        const stack: number[] = [root];
        while(stack.length > 0) {
            const index = stack.pop() ?? panic('missing traversal');
            if(index < 0) {
                const leaving = this.get(-index - 1);
                if(leaving.kind === 'group' || leaving.kind === 'conditionalGroup') {
                    groups.pop();
                    const parent = groups[groups.length - 1];
                    if(leaving.broken && parent !== undefined && this.get(parent).kind !== 'conditionalGroup')
                        this.get(parent).broken = true;
                }
                continue;
            }
            const document = this.get(index);
            stack.push(-index - 1);
            if(document.kind === 'breakParent') {
                const parent = groups[groups.length - 1];
                if(parent !== undefined && this.get(parent).kind !== 'conditionalGroup') this.get(parent).broken = true;
            }
            if(document.kind === 'group' || document.kind === 'conditionalGroup') {
                groups.push(index);
                if(visited.has(index)) continue;
                visited.add(index);
            }
            for(let child = document.parts.length - 1; child >= 0; child--)
                stack.push(document.parts[child] ?? panic('missing doc child'));
        }
    }
    fits(
        next: CommandInterface,
        rest: readonly CommandInterface[],
        width: number,
        suffix: boolean,
        modes: ReadonlyMap<string, number>,
        mustFlat: boolean,
    ): boolean {
        let restIndex = rest.length;
        const commands: CommandInterface[] = [next];
        let pending = false;
        let trailingWidth = 0;
        let hasSuffix = suffix;
        let remaining = width;
        while(remaining >= 0) {
            if(commands.length === 0) {
                if(restIndex === 0) return true;
                restIndex--;
                commands.push(rest[restIndex] ?? panic('missing rest command'));
                continue;
            }
            const current = commands.pop() ?? panic('missing fits command');
            const document = this.get(current.doc);

            if(document.kind === 'text') {
                if(document.text !== '') {
                    if(pending) {
                        trailingWidth++;
                        remaining--;
                        pending = false;
                    }
                    trailingWidth = trailing(document.text, trailingWidth);
                    remaining -= stringWidth(document.text);
                }
            }
            else if(document.kind === 'trim') {
                remaining += trailingWidth;
                trailingWidth = 0;
            }
            else if(document.kind === 'group' || document.kind === 'conditionalGroup') {
                if(mustFlat && document.broken) return false;
                const mode = document.broken ? 1 : current.mode;
                const contents =
                    document.parts[
                        document.kind === 'conditionalGroup' && mode === 1 ? document.parts.length - 1 : 0
                    ] ?? panic('missing group contents');
                commands.push({ doc: contents, indent: 0, mode, offset: 0 });
            }
            else if(document.kind === 'ifBreak') {
                const mode = document.key === '' ? current.mode : (modes.get(document.key) ?? 2);
                commands.push({
                    doc: document.parts[mode === 1 ? 0 : 1] ?? panic('missing ifBreak branch'),
                    indent: 0,
                    mode: current.mode,
                    offset: 0,
                });
            }
            else if(
                document.kind === 'line' ||
                document.kind === 'softline' ||
                document.kind === 'hardlineWithoutBreakParent' ||
                document.kind === 'literallineWithoutBreakParent'
            ) {
                if(
                    current.mode === 1 ||
                    document.kind === 'hardlineWithoutBreakParent' ||
                    document.kind === 'literallineWithoutBreakParent'
                )
                    return true;
                if(document.kind !== 'softline') pending = true;
            }
            else if(document.kind === 'lineSuffix') {
                hasSuffix = true;
            }
            else if(document.kind === 'lineSuffixBoundary') {
                if(hasSuffix) return false;
            }
            else if(document.kind !== 'breakParent') {
                const start = document.kind === 'fill' ? current.offset : 0;
                for(let index = document.parts.length - 1; index >= start; index--)
                    commands.push({
                        doc: document.parts[index] ?? panic('missing fits child'),
                        indent: 0,
                        mode: current.mode,
                        offset: 0,
                    });
            }
        }
        return false;
    }
    print(root: number): string {
        this.propagate(root);
        const modes = new Map<string, number>();
        const indents: IndentInterface[] = [{ value: '', length: 0, queue: [], strings: [], root: 0 }];
        const commands: CommandInterface[] = [{ doc: root, indent: 0, mode: 1, offset: 0 }];
        const suffixes: CommandInterface[] = [];
        const output: string[] = [];
        let column = 0;
        let remeasure = false;
        while(commands.length > 0) {
            const current = commands.pop() ?? panic('missing print command');
            const document = this.get(current.doc);

            if(document.kind === 'text') {
                if(document.text !== '') {
                    output.push(document.text);
                    if(commands.length > 0) column += stringWidth(document.text);
                }
            }
            else if(
                document.kind === 'indent' ||
                document.kind === 'alignWidth' ||
                document.kind === 'alignString' ||
                document.kind === 'markAsRoot' ||
                document.kind === 'dedentToRoot'
            ) {
                const base = indents[current.indent] ?? panic('missing indentation');
                let indent: number;
                if(document.kind === 'dedentToRoot') indent = base.root;
                else if(document.kind === 'markAsRoot') {
                    indents.push({ ...base, root: current.indent });
                    indent = indents.length - 1;
                }
                else {
                    const queue = base.queue.slice();
                    const strings = base.strings.slice();
                    if(document.kind === 'alignWidth' && document.number < 0) {
                        queue.pop();
                        strings.pop();
                    }
                    else if(document.kind !== 'alignWidth' || document.number !== 0) {
                        queue.push(
                            document.kind === 'indent' ? -1 : document.kind === 'alignString' ? -2 : document.number,
                        );
                        strings.push(document.text);
                    }
                    let value = '';
                    let length = 0;
                    let tabs = 0;
                    let spaces = 0;
                    for(let index = 0; index < queue.length; index++) {
                        const command = queue[index] ?? panic('missing indent instruction');
                        if(command >= 0) {
                            tabs++;
                            spaces += command;
                            continue;
                        }
                        const pending = this.settings.useTabs ? '\t'.repeat(tabs) : ' '.repeat(spaces);
                        value += pending;
                        length += this.settings.useTabs ? tabs * this.settings.tabWidth : spaces;
                        tabs = 0;
                        spaces = 0;
                        const addition =
                            command === -1
                                ? this.settings.useTabs
                                    ? '\t'
                                    : ' '.repeat(this.settings.tabWidth)
                                : (strings[index] ?? panic('missing align string'));
                        value += addition;
                        length += command === -1 ? this.settings.tabWidth : addition.length;
                    }
                    value += ' '.repeat(spaces);
                    length += spaces;
                    indents.push({ value, length, queue, strings, root: base.root });
                    indent = indents.length - 1;
                }
                commands.push({
                    doc: document.parts[0] ?? panic('missing indented doc'),
                    indent,
                    mode: current.mode,
                    offset: 0,
                });
            }
            else if(document.kind === 'trim') {
                column -= this.trim(output);
            }
            else if(document.kind === 'group' || document.kind === 'conditionalGroup') {
                let mode: number;
                let contents = document.parts[0] ?? panic('missing group contents');
                if(current.mode === 2 && !remeasure) mode = document.broken ? 1 : 2;
                else {
                    remeasure = false;
                    const flat: CommandInterface = { doc: contents, indent: current.indent, mode: 2, offset: 0 };
                    if(
                        !document.broken &&
                        this.fits(flat, commands, this.settings.printWidth - column, suffixes.length > 0, modes, false)
                    )
                        mode = 2;
                    else {
                        mode = 1;
                        if(document.kind === 'conditionalGroup') {
                            contents = document.parts[document.parts.length - 1] ?? panic('missing last state');
                            if(!document.broken) {
                                for(let index = 1; index < document.parts.length - 1; index++) {
                                    const state = document.parts[index] ?? panic('missing group state');
                                    if(
                                        this.fits(
                                            { doc: state, indent: current.indent, mode: 2, offset: 0 },
                                            commands,
                                            this.settings.printWidth - column,
                                            suffixes.length > 0,
                                            modes,
                                            false,
                                        )
                                    ) {
                                        contents = state;
                                        mode = 2;
                                        break;
                                    }
                                }
                            }
                        }
                    }
                }
                commands.push({ doc: contents, indent: current.indent, mode, offset: 0 });
                if(document.key !== '') modes.set(document.key, mode);
            }
            else if(document.kind === 'fill') {
                const length = document.parts.length - current.offset;
                if(length > 0) {
                    const content = document.parts[current.offset] ?? panic('missing fill content');
                    const fitsContent = this.fits(
                        { doc: content, indent: current.indent, mode: 2, offset: 0 },
                        [],
                        this.settings.printWidth - column,
                        suffixes.length > 0,
                        modes,
                        true,
                    );
                    let contentMode = fitsContent ? 2 : 1;
                    let separatorMode = contentMode;
                    if(length > 2) {
                        const pair = this.concat(document.parts.slice(current.offset, current.offset + 3));
                        const fitsPair = this.fits(
                            { doc: pair, indent: current.indent, mode: 2, offset: 0 },
                            [],
                            this.settings.printWidth - column,
                            suffixes.length > 0,
                            modes,
                            true,
                        );
                        contentMode = fitsPair || fitsContent ? 2 : 1;
                        separatorMode = fitsPair ? 2 : 1;
                        commands.push({
                            doc: current.doc,
                            indent: current.indent,
                            mode: current.mode,
                            offset: current.offset + 2,
                        });
                    }
                    if(length > 1)
                        commands.push({
                            doc: document.parts[current.offset + 1] ?? panic('missing fill separator'),
                            indent: current.indent,
                            mode: separatorMode,
                            offset: 0,
                        });
                    commands.push({ doc: content, indent: current.indent, mode: contentMode, offset: 0 });
                }
            }
            else if(document.kind === 'ifBreak' || document.kind === 'indentIfBreak') {
                const mode = document.key === '' ? current.mode : (modes.get(document.key) ?? 0);
                if(mode === 1 || mode === 2) {
                    let contents =
                        document.parts[document.kind === 'ifBreak' && mode === 2 ? 1 : 0] ??
                        panic('missing break choice');
                    if(
                        document.kind === 'indentIfBreak' &&
                        ((mode === 1 && !document.broken) || (mode === 2 && document.broken))
                    )
                        contents = this.indent(contents);
                    commands.push({ doc: contents, indent: current.indent, mode: current.mode, offset: 0 });
                }
            }
            else if(document.kind === 'lineSuffix') {
                suffixes.push({
                    doc: document.parts[0] ?? panic('missing suffix'),
                    indent: current.indent,
                    mode: current.mode,
                    offset: 0,
                });
            }
            else if(document.kind === 'lineSuffixBoundary') {
                if(suffixes.length > 0)
                    commands.push({
                        doc: this.add('hardlineWithoutBreakParent', []),
                        indent: current.indent,
                        mode: current.mode,
                        offset: 0,
                    });
            }
            else if(
                document.kind === 'line' ||
                document.kind === 'softline' ||
                document.kind === 'hardlineWithoutBreakParent' ||
                document.kind === 'literallineWithoutBreakParent'
            ) {
                const hard =
                    document.kind === 'hardlineWithoutBreakParent' || document.kind === 'literallineWithoutBreakParent';
                if(current.mode === 2 && !hard) {
                    if(document.kind !== 'softline') {
                        output.push(' ');
                        column++;
                    }
                }
                else {
                    if(current.mode === 2) remeasure = true;
                    if(suffixes.length > 0) {
                        commands.push(current);
                        for(let index = suffixes.length - 1; index >= 0; index--)
                            commands.push(suffixes[index] ?? panic('missing pending suffix'));
                        suffixes.splice(0);
                    }
                    else {
                        const indent = indents[current.indent] ?? panic('missing line indent');
                        const target =
                            document.kind === 'literallineWithoutBreakParent'
                                ? (indents[indent.root] ?? panic('missing root indent'))
                                : indent;
                        if(document.kind !== 'literallineWithoutBreakParent') this.trim(output);
                        output.push('\n');
                        output.push(target.value);
                        column = target.length;
                    }
                }
            }
            else if(document.kind !== 'breakParent') {
                for(let index = document.parts.length - 1; index >= 0; index--)
                    commands.push({
                        doc: document.parts[index] ?? panic('missing print child'),
                        indent: current.indent,
                        mode: current.mode,
                        offset: 0,
                    });
            }
            if(commands.length === 0 && suffixes.length > 0) {
                for(let index = suffixes.length - 1; index >= 0; index--)
                    commands.push(suffixes[index] ?? panic('missing final suffix'));
                suffixes.splice(0);
            }
        }
        return output.join('');
    }
    /** @mutates output removes trailing whitespace from the printer's own output chunks */
    trim(output: string[]): number {
        let removed = 0;
        while(output.length > 0) {
            const last = output[output.length - 1] ?? panic('missing output');
            const count = trailing(last, 0);
            removed += count;
            if(count === last.length) {
                output.pop();
                continue;
            }
            if(count > 0) output[output.length - 1] = last.slice(0, -count);
            break;
        }
        return removed;
    }
}
