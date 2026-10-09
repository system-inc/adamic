// cohere CSS printer-postcss, declaration/rule printers, and value-group printers.
import { Tree } from './tree.ts';
import { compose } from './compose.ts';
import { byteSlice } from './convert.ts';
import { utf8Length } from 'adamic';
import { quote } from '../selector/nodes.ts';
import { adjustStrings, printString } from '../cssstrings/strings.ts';
import { adjustNumbers } from '../cssnumbers/numbers.ts';
import { Documents } from './print_doc.ts';
import { Node, Path, lower, maybeLower, inlineLast, emptyBefore, control, placeholder, operator, word, valueWord, inlineValue, pair, parenPair, parens, mapItem, breakList, precedeSoft, hasNewline, nextEmpty } from './print_utilities.ts';
export type PrintOptions = { readonly width: number; readonly tabWidth: number; readonly tabs: boolean; readonly single: boolean; readonly trailingComma: string };
export const defaults: PrintOptions = {width: 80, tabWidth: 2, tabs: false, single: false, trailingComma: 'all'};
const knownUnits: readonly string[] = 'em rem ex rex cap rcap ch rch ic ric lh rlh vw svw lvw dvw vh svh lvh dvh vi svi lvi dvi vb svb lvb dvb vmin svmin lvmin dvmin vmax svmax lvmax dvmax cm mm q in pt pc px deg grad rad turn s ms hz khz dpi dpcm dppx x cqw cqh cqi cqb cqmin cqmax fr'.split(' ');
export class Printer {
    readonly tree: Tree;
    readonly text: string;
    readonly options: PrintOptions;
    readonly scss: boolean;
    readonly docs = new Documents();
    constructor(tree: Tree, text: string, options: PrintOptions, scss: boolean) { this.tree = tree; this.text = text; this.options = options; this.scss = scss; }
    adjusted(value: string): string { return adjustNumbers(adjustStrings(value, this.options.single)); }
    textOf(n: Node): string { return byteSlice(this.text, n.start(), n.end()); }
    list(path: Path, key: string): number[] {
        const printed: number[] = [];
        for(let i = 0; i < path.node().list(key).length; i++) printed.push(this.print(path.child(key, i)));
        return printed;
    }
    sequence(path: Path): number {
        const d = this.docs; const parts: number[] = []; const nodes = path.node().list('nodes');
        for(let i = 0; i < nodes.length; i++) {
            const child = path.child('nodes', i); const n = child.node(); const previous = child.sibling(-1); const next = child.sibling(1);
            parts.push(previous.type() === 'css-comment' && previous.string('text').trim() === 'prettier-ignore' ? d.text(this.textOf(n)) : this.print(child));
            if(i === nodes.length - 1) continue;
            if((next.type() === 'css-comment' && !hasNewline(this.text, next.start(), true) && n.type() !== 'front-matter') || (next.type() === 'css-atrule' && next.string('name') === 'else' && n.type() !== 'css-comment')) parts.push(d.text(' '));
            else { parts.push(d.hard()); if(nextEmpty(this.text, n.end()) && n.type() !== 'front-matter') parts.push(d.hard()); }
        }
        return d.concat(parts);
    }
    namespace(n: Node): number {
        const d = this.docs;
        return !n.truth('namespace') ? d.text('') : d.text((n.data().booleans.get('namespace') ? '' : n.string('namespace').trim()) + '|');
    }
    print(path: Path): number {
        const n = path.node(); const p = path.up(1); const d = this.docs; const value = n.string('value');
        const child = (key: string) => { if(n.data().strings.has(key)) throw new Error('Unexpected PostCSS node type: '+quote('')+'.'); return this.print(path.child(key)); };
        if(n.index < 0) return d.text('');
        if(n.data().literal) return d.text(n.data().literalValue);
        switch(n.type()) {
            case 'front-matter': return ['yaml','toml'].includes(n.string('language')) && n.string('value').trim() === '' ? d.concat([d.text(n.string('startDelimiter')+n.string('explicitLanguage')),d.hard(),d.text(n.string('endDelimiter'))]) : d.text(n.string('raw'));
            case 'css-root': {
                let after = n.raw('after').trim(); if(after.startsWith(';')) after = after.slice(1).trim();
                return d.concat([n.child('frontMatter').index >= 0 ? d.concat([child('frontMatter'), d.hard(), n.list('nodes').length > 0 ? d.hard() : d.text('')]) : d.text(''), this.sequence(path), d.text(after === '' ? '' : ' ' + after), n.list('nodes').length > 0 ? d.hard() : d.text('')]);
            }
            case 'css-comment': return d.text(n.truth('inline') || n.child('raws').truth('inline') ? this.textOf(n).trimEnd() : this.textOf(n));
            case 'css-rule': {
                const selector = n.child('selector');
                let block = d.text(';');
                if(n.truth('nodes')) block = d.concat([selector.type() === 'selector-unknown' && inlineLast(selector.string('value')) ? d.line() : d.text(selector.index >= 0 ? ' ' : ''), d.text('{'), n.list('nodes').length > 0 ? d.indent(d.concat([d.hard(), this.sequence(path)])) : d.text(''), d.hard(), d.text('}'), d.text(/^@[^\n\r\u2028\u2029]+:[^\n\r\u2028\u2029]*$/.test(selector.string('value')) ? ';' : '')]);
                return d.concat([child('selector'), d.text(n.truth('important') ? ' !important' : ''), block]);
            }
            case 'css-decl': return this.declaration(path);
            case 'css-atrule': return this.atRule(path);
            case 'media-query-list': {
                const parts: number[] = [];
                for(let i = 0; i < n.list('nodes').length; i++) { const c = path.child('nodes', i); if(c.node().type() === 'media-query' && c.node().string('value') === '') continue; parts.push(this.print(c)); }
                return d.group(d.indent(d.join(d.line(), parts)));
            }
            case 'media-query': return d.concat([d.join(d.text(' '), this.list(path, 'nodes')), d.text(path.last() ? '' : ',')]);
            case 'media-type': case 'media-value': return d.text(this.adjusted(value));
            case 'media-feature-expression': return n.truth('nodes') ? d.concat([d.text('('), ...this.list(path,'nodes'), d.text(')')]) : d.text(value);
            case 'media-feature': return d.text(maybeLower(adjustStrings(value.replace(/ +/g, ' '), this.options.single)));
            case 'media-colon': return d.text(value + ' ');
            case 'media-keyword': return d.text(adjustStrings(value, this.options.single));
            case 'media-url': return d.text(adjustStrings(value.replace(/^url\(\s+/i, 'url(').replace(/\s+\)$/, ')'), this.options.single));
            case 'media-unknown': return d.text(value);
            case 'selector-root': return d.group(d.concat([path.insideRule(['custom-selector']) ? d.concat([d.text(path.ancestor('css-atrule').string('customSelector')), d.line()]) : d.text(''), d.join(d.concat([d.text(','), path.insideRule(['extend','custom-selector','nest']) ? d.line() : d.hard()]), this.list(path,'nodes'))]));
            case 'selector-selector': { const body = d.concat(this.list(path,'nodes')); return d.group(n.list('nodes').length > 2 ? d.indent(body) : body); }
            case 'selector-comment': case 'selector-nesting': return d.text(value);
            case 'selector-string': return d.text(adjustStrings(value, this.options.single));
            case 'selector-tag': { let adjusted = value; if(path.sibling(-1).type() !== 'selector-nesting') { if(path.ancestor('css-atrule').string('name').toLowerCase().endsWith('keyframes') && ['from','to'].includes(lower(value))) adjusted = lower(value); adjusted = adjustNumbers(adjusted); } return d.concat([this.namespace(n), d.text(adjusted)]); }
            case 'selector-id': return d.text('#' + value);
            case 'selector-class': return d.text('.' + this.adjusted(value));
            case 'selector-attribute': {
                let attr = adjustStrings(value.trim(), this.options.single); let flag = ''; const match = /^([^\n\r\u2028\u2029]+?)\s+([a-z])$/i.exec(attr);
                if(match !== null) { attr = match[1] ?? ''; flag = match[2] ?? ''; }
                if(attr !== '' && !attr.includes('"') && !attr.includes("'")) { const q = this.options.single ? "'" : '"'; attr = q + attr + q; }
                if(flag !== '') attr += ' ' + flag;
                const parts: number[] = []; const lines = attr.split('\n'); for(let i = 0; i < lines.length; i++) { if(i > 0) parts.push(d.literal()); parts.push(d.text(lines[i] ?? '')); }
                return d.concat([d.text('['), this.namespace(n), d.text(n.string('attribute').trim() + n.string('operator')), d.concat(parts), d.text(n.truth('insensitive') ? ' i]' : ']')]);
            }
            case 'selector-combinator': {
                if(['+','>','~','>>>'].includes(value)) return d.concat([p.type() === 'selector-selector' && p.list('nodes')[0] === n.index ? d.text('') : d.line(), d.text(value), d.text(path.last() ? '' : ' ')]);
                const adjusted = this.adjusted(value.trim()); return d.concat([value.trimStart().startsWith('(') ? d.line() : d.text(''), adjusted === '' ? d.line() : d.text(adjusted)]);
            }
            case 'selector-universal': return d.concat([this.namespace(n), d.text(value)]);
            case 'selector-pseudo': return d.concat([d.text(maybeLower(value)), n.list('nodes').length > 0 ? d.group(d.concat([d.text('('), d.indent(d.concat([d.soft(), d.join(d.concat([d.text(','), d.line()]), this.list(path,'nodes'))])), d.soft(), d.text(')')])) : d.text('')]);
            case 'selector-unknown': {
                if(path.ancestor('css-rule').truth('isScssNestedProperty')) return d.text(this.adjusted(maybeLower(value)));
                if(p.raw('selector') !== '') return d.text(byteSlice(this.text, p.start(), p.start() + utf8Length(p.raw('selector'))).trim());
                if(p.type() === 'value-paren_group' && path.up(2).type() === 'value-func' && path.up(2).string('value') === 'selector') {
                    const text = byteSlice(this.text, p.child('open').end() + 1, p.child('close').start()).trim(); return inlineLast(text) ? d.concat([d.add('break'), d.text(text)]) : d.text(text);
                }
                return d.text(value);
            }
            case 'value-root': case 'value-value': return child('group');
            case 'value-comment': return n.truth('inline') ? d.suffix(d.text(this.textOf(n).trimEnd())) : d.text(this.textOf(n));
            case 'value-comma_group': return this.comma(path);
            case 'value-paren_group': return this.parentheses(path);
            case 'value-func': return d.concat([d.text(value), d.text(path.insideRule(['supports']) && ['not','and','or'].includes(lower(value)) ? ' ' : ''), child('group')]);
            case 'value-paren': case 'value-operator': case 'value-unicode-range': case 'value-unknown': return d.text(value);
            case 'value-number': {
                let number = value.length === 1 ? value : value.toLowerCase()
                    .replace(/^([+-]?[\d.]+e)(?:\+|(-))?0*(\d)/, '$1$2$3')
                    .replace(/^([+-]?[\d.]+)e[+-]?0+$/, '$1')
                    .replace(/^([+-])?\./, '$10.')
                    .replace(/(\.\d+?)0+(e|$)/, '$1$2')
                    .replace(/\.(e|$)/, '$1');
                number = number.replace(/\.0($|e)/, '$1');
                const rawUnit = n.string('unit');
                const lowerUnit = lower(rawUnit);
                const known = knownUnits;
                const unit = !known.includes(lowerUnit) ? rawUnit : lowerUnit === 'q' ? 'Q' : lowerUnit === 'hz' ? 'Hz' : lowerUnit === 'khz' ? 'kHz' : lowerUnit;
                return d.text(number + unit);
            }
            case 'value-word': return d.text((n.truth('isColor') && n.truth('isHex')) || ['initial','inherit','unset','revert'].includes(lower(value)) ? lower(value) : value);
            case 'value-colon': return d.group(d.concat([d.text(value), path.sibling(-1).string('value').endsWith('\\') || path.insideFunc('url') ? d.text('') : d.line()]));
            case 'value-string': { const q = n.raw('quote'); return d.text(printString(q + value + q, this.options.single)); }
            case 'value-atword': return d.text('@' + value);
            default: throw new Error('Unexpected PostCSS node type: ' + quote(n.type()) + '.');
        }
    }
    declaration(path: Path): number {
        const n = path.node(); const p = path.up(1); const d = this.docs;
        const between = n.raw('between'); const trimmed = between.trim(); const colon = trimmed === ':';
        const stringValue = n.data().strings.has('value');
        let value = stringValue ? d.text(n.string('value')) : this.print(path.child('value'));
        if(n.child('value').type() === 'value-root' && n.child('value').child('group').type() === 'value-value' && lower(n.string('prop')) === 'composes') value = d.removeLines(value);
        if(!colon && inlineLast(trimmed) && !breakList(path.child('value').child('group').child('group'))) value = d.indent(d.concat([d.hard(), d.dedent(value)]));
        const prop = (p.type() === 'css-atrule' && p.truth('variable')) || path.icss() ? n.string('prop') : maybeLower(n.string('prop'));
        let space = ' ';
        if(n.truth('extend') || (stringValue && /^ *$/.test(n.string('value'))) || (!(between.endsWith(' ') && colon) && n.truth('isNested') && (placeholder(n.child('value').child('group').child('group')) || placeholder(new Node(this.tree, n.child('value').child('group').child('group').list('groups')[0] ?? -1))))) space = '';
        const flag = (name: string, pattern: RegExp, suffix: string) => n.raw(name) !== '' ? n.raw(name).replace(pattern, suffix) : n.truth(name) ? suffix : '';
        let ending = d.text(';');
        if(n.truth('nodes')) ending = d.concat([d.text(' {'), n.list('nodes').length > 0 ? d.indent(d.concat([d.soft(), this.sequence(path)])) : d.text(''), d.soft(), d.text('}')]);
        else if(n.string('prop').startsWith('@prettier-placeholder') && !p.child('raws').truth('semicolon') && byteSlice(this.text, n.end() - 1, n.end()) !== ';') ending = d.text('');
        return d.concat([d.text(n.raw('before').replace(/[\s;]/g,'')), d.text(prop), d.text(trimmed.startsWith('//') ? ' ' : ''), d.text(trimmed), d.text(space), value, d.text(flag('important', /\s*!\s*important/i, ' !important')), d.text(flag('scssDefault', /\s*!default/i, ' !default')), d.text(flag('scssGlobal', /\s*!global/i, ' !global')), ending]);
    }
    atRule(path: Path): number {
        const n = path.node(); const p = path.up(1); const d = this.docs; const name = n.string('name'); const params = n.child('params'); const selector = n.child('selector');
        const template = name.startsWith('prettier-placeholder'); const detached = /^\(\s*\)$/.test(n.raw('params'));
        const printedName = detached || name.endsWith(':') || template ? name : maybeLower(name);
        let printedParams = d.text('');
        if(n.truth('params')) {
            let leading = d.text(detached ? '' : ' ');
            if(template) { const after = n.raw('afterName'); if(after === '') leading = d.text(''); else if(!name.endsWith(':')) { if(/^\s*\n\s*\n/.test(after)) leading = d.concat([d.hard(),d.hard()]); else if(/^\s*\n/.test(after)) leading = d.hard(); } }
            printedParams = d.concat([leading, n.data().strings.has('params') ? d.text(n.string('params')) : this.print(path.child('params'))]);
        }
        const printedSelector = selector.index < 0 ? d.text('') : d.indent(d.concat([d.text(' '), this.print(path.child('selector'))]));
        let printedValue = d.text(name === 'else' ? ' ' : '');
        if(n.truth('value')) { let after = d.text(''); if(control(n,this.scss)) { const group = n.child('value').child('group').child('group'); after = group.type() === 'value-paren_group' && !group.data().nulls.has('open') && !group.data().nulls.has('close') ? d.text(' ') : d.line(); } printedValue = d.group(d.concat([d.text(' '), this.print(path.child('value')), after])); }
        let ending = d.text(';');
        if(n.truth('nodes')) {
            let opening = d.text(control(n,this.scss) ? '' : ' ');
            if(!control(n,this.scss) && ((selector.index >= 0 && !selector.truth('nodes') && selector.data().strings.has('value') && inlineLast(selector.string('value'))) || (selector.index < 0 && n.data().strings.has('params') && inlineLast(n.string('params'))))) opening = d.line();
            ending = d.concat([opening,d.text('{'), n.list('nodes').length > 0 ? d.indent(d.concat([d.soft(), this.sequence(path)])) : d.text(''), d.soft(), d.text('}')]);
        } else if((template && !p.child('raws').truth('semicolon') && byteSlice(this.text,n.end()-1,n.end()) !== ';') || (name === 'import' && params.type() === 'value-unknown' && params.string('value').endsWith(';'))) ending = d.text('');
        return d.concat([d.text('@'+printedName),printedParams,printedSelector,printedValue,ending]);
    }
    trailing(path: Path): number {
        const d = this.docs; const n = path.node(); const parent = path.up(1); const grand = path.up(2);
        if(grand.type() === 'value-func' && lower(grand.string('value')) === 'var' && n.truth('source') && byteSlice(this.text,n.start(),parent.child('close').start()).trimEnd().endsWith(',')) return d.text(',');
        if(n.type() !== 'value-comment' && !(n.type() === 'value-comma_group' && n.list('groups').every(i => new Node(this.tree,i).type() === 'value-comment')) && ['es5','all'].includes(this.options.trailingComma) && mapItem(path.parent(), this.scss)) return d.ifBreak(d.text(','));
        return d.text('');
    }
    parentheses(path: Path): number {
        const n = path.node(); const p = path.up(1); const d = this.docs; const groups = n.list('groups'); const printed = this.list(path,'groups'); const first = new Node(this.tree,groups[0] ?? -1);
        if(p.type() === 'value-func' && lower(p.string('value')) === 'url' && (groups.length === 1 || (first.type() === 'value-comma_group' && new Node(this.tree,first.list('groups')[0] ?? -1).type() === 'value-word' && new Node(this.tree,first.list('groups')[0] ?? -1).string('value').startsWith('data:')))) return d.concat([n.child('open').index < 0 ? d.text('') : this.print(path.child('open')), d.join(d.text(','), printed), n.child('close').index < 0 ? d.text('') : this.print(path.child('close'))]);
        if(n.child('open').index < 0) {
            const joined: number[] = []; for(let i = 0; i < printed.length; i++) { joined.push(d.concat([printed[i] ?? 0, i < printed.length - 1 ? d.text(',') : d.text('')])); }
            const force = breakList(path);
            if(force) return d.indent(d.concat([d.hard(), d.join(d.hard(),joined)]));
            const fill: number[] = []; for(let i = 0; i < joined.length; i++) { if(i > 0) fill.push(d.line()); fill.push(joined[i] ?? 0); }
            const parts: number[] = precedeSoft(path) ? [d.soft()] : []; parts.push(d.fill(fill));
            return d.indent(d.group(d.concat(parts)));
        }
        const parts: number[] = [];
        for(let i = 0; i < groups.length; i++) {
            const c = path.child('groups',i); const child = c.node(); const childGroups = child.list('groups'); let document = printed[i] ?? 0;
            if(pair(child) && new Node(this.tree,childGroups[0] ?? -1).type() !== 'value-paren_group' && childGroups.length > 2 && new Node(this.tree,childGroups[2] ?? -1).type() === 'value-paren_group') {
                const outer = d.at(document); const middle = d.at(outer.parts[0] ?? 0);
                if(outer.kind === 'group' && middle.kind === 'indent' && d.at(middle.parts[0] ?? 0).kind === 'fill') document = d.group(d.dedent(document));
            }
            const one: number[] = [document, i === groups.length - 1 ? this.trailing(c) : d.text(',')];
            if(i < groups.length - 1 && child.type() === 'value-comma_group' && childGroups.length > 0) { let last = new Node(this.tree, childGroups[childGroups.length - 1] ?? -1); if(!last.truth('source') && last.child('close').index >= 0) last = last.child('close'); if(last.truth('source') && nextEmpty(this.text,last.end())) one.push(d.hard()); }
            parts.push(d.concat(one));
        }
        const pg = p.list('groups'); const i = pg.indexOf(n.index); const isKey = pair(p) && i >= 0 && new Node(this.tree,pg[i+1] ?? -1).type() === 'value-colon';
        const previous = new Node(this.tree,pg[i-1] ?? -1); const config = parens(n) && groups.every(g => new Node(this.tree,g).type() === 'value-comma_group') && p.type() === 'value-comma_group' && valueWord(previous,'with');
        const document = d.group(d.concat([n.child('open').index < 0 ? d.text('') : this.print(path.child('open')), d.indent(d.concat([d.soft(),d.join(d.line(),parts)])), d.soft(),d.boundary(),n.child('close').index < 0 ? d.text('') : this.print(path.child('close'))]), config || (mapItem(path,this.scss) && !isKey));
        return config || isKey ? d.dedent(document) : document;
    }
    comma(path: Path): number {
        const n = path.node(); const parent = path.up(1); const grand = path.up(2); const d = this.docs; const groups = n.list('groups'); const printed = this.list(path,'groups'); const prop = lower(path.ancestor('css-decl').string('prop')); const at = path.ancestor('css-atrule');
        const grid = prop !== '' && parent.type() === 'value-value' && (prop === 'grid' || prop.startsWith('grid-template')); const directive = control(at,this.scss);
        let parts: number[] = [d.text('')]; const append = (doc: number) => { const last = parts.length - 1; parts[last] = d.concat([parts[last] ?? 0,doc]); };
        const atIndex = (i: number) => new Node(this.tree,groups[i] ?? -1);
        const color = (node: Node) => node.type() === 'value-func' && ['red','green','blue','alpha','a','rgb','hue','h','saturation','s','lightness','l','whiteness','w','blackness','b','tint','shade','blend','blenda','contrast','hsl','hsla','hwb','hwba'].includes(lower(node.string('value')));
        const fontSize = (node: Node) => node.type() === 'value-number' || (node.type() === 'value-func' && (['var','calc','min','max','clamp'].includes(lower(node.string('value'))) || lower(node.string('value')).startsWith('--')));
        let interpolation = false; let didBreak = false;
        for(let i = 0; i < groups.length; i++) {
            const prev = atIndex(i-1); const current = atIndex(i); const next = atIndex(i+1); const nextNext = atIndex(i+2); const cv = current.string('value'); const nv = next.string('value');
            if(inlineValue(current) && next.index < 0) { append(d.suffix(d.concat([d.text(' '), printed[i] ?? 0]))); continue; }
            append(printed[i] ?? 0);
            if(path.insideFunc('url')) { if(operator(next,'+') || operator(current,'+')) append(d.text(' ')); continue; }
            if(path.insideRule(['forward']) && current.type() === 'value-word' && cv !== '' && valueWord(prev,'as') && operator(next,'*')) continue;
            if(path.insideRule(['utility']) && current.type() === 'value-word' && operator(next,'*')) continue;
            if(next.index < 0) continue;
            if(this.scss && valueWord(next,';') && parent.type() === 'value-paren_group' && path.key() === 'groups' && grand.type() === 'value-func' && path.key(1) === 'group' && grand.string('value') === 'if') continue;
            if(current.type() === 'value-word' && placeholder(next) && current.end() === next.start()) continue;
            if(current.type() === 'value-string' && current.truth('quoted')) { const open = cv.lastIndexOf('#{'); const close = cv.lastIndexOf('}'); if(open >= 0 && close >= 0) interpolation = open > close; else if(open >= 0) interpolation = true; else if(close >= 0) interpolation = false; }
            if(interpolation || current.type() === 'value-colon' || next.type() === 'value-colon') continue;
            if(current.type() === 'value-atword' && (cv === '' || cv.endsWith('['))) continue;
            if(next.type() === 'value-word' && nv.startsWith(']')) continue;
            if(cv === '~') continue;
            if(current.type() !== 'value-string' && cv !== '' && cv.includes('\\') && next.type() !== 'value-comment') continue;
            if(prev.string('value') !== '' && prev.string('value').indexOf('\\') === prev.string('value').length - 1 && operator(current,'/')) continue;
            if(cv === '\\') continue;
            if(cv === '$$' && current.type() === 'value-func' && next.type() === 'value-word' && !next.child('raws').truth('before')) continue;
            if(valueWord(current,'#') || valueWord(current,'{') || valueWord(next,'}') || (valueWord(next,'{') && emptyBefore(next)) || (valueWord(current,'}') && emptyBefore(next))) continue;
            if(cv === '--' && valueWord(next,'#')) continue;
            const math = operator(current); const nextMath = operator(next);
            if(((math && valueWord(next,'#')) || (nextMath && valueWord(current,'}'))) && emptyBefore(next)) continue;
            if(operator(next,'+') && path.insideFunc('type') && emptyBefore(next)) continue;
            if(prev.index < 0 && operator(current,'/')) continue;
            if(path.insideFunc('calc') && (operator(current,'+-') || operator(next,'+-')) && emptyBefore(next)) continue;
            if(this.scss && operator(current,'-') && next.type() === 'value-func' && current.end() !== next.start()) { append(d.text(' ')); continue; }
            const adjuster = operator(current,'+-') && i === 0 && (next.type() === 'value-number' || next.truth('isHex')) && color(grand) && !emptyBefore(next);
            const before = nextNext.type() === 'value-func' || word(nextNext) || current.type() === 'value-func' || word(current);
            const after = next.type() === 'value-func' || word(next) || prev.type() === 'value-func' || word(prev);
            if(!(operator(next,'*') || operator(current,'*')) && !path.insideFunc('calc') && !adjuster && ((operator(next,'/') && !before) || (operator(current,'/') && !after) || (operator(next,'+') && !before) || (operator(current,'+') && !after) || operator(next,'-') || operator(current,'-')) && (emptyBefore(next) || (math && (prev.index < 0 || operator(prev))))) continue;
            if(this.scss && operator(current,'-') && parens(next) && current.end() === next.child('open').start()) continue;
            if(inlineValue(current)) { parts.push(parent.type() === 'value-paren_group' ? d.dedent(d.hard()) : d.hard()); parts.push(d.text('')); continue; }
            if(directive && ((next.type() === 'value-word' && ['==','!=','<','>','<=','>=','and','or','not'].includes(nv)) || valueWord(current,'in') || (current.type() === 'value-word' && ['from','through','end'].includes(cv)))) { append(d.text(' ')); continue; }
            if(lower(at.string('name')) === 'namespace') { append(d.text(' ')); continue; }
            if(grid) { if(current.truth('source') && next.truth('source') && current.line() !== next.line()) { parts.push(d.hard()); parts.push(d.text('')); didBreak = true; } else append(d.text(' ')); continue; }
            if(prop === 'font' || prop.startsWith('--')) { if(operator(next,'/') && emptyBefore(next) && fontSize(current)) continue; if(operator(current,'/') && emptyBefore(current) && fontSize(prev)) continue; }
            if(nextMath) { append(d.text(' ')); continue; }
            if(nv === '...') continue;
            if(placeholder(current) && placeholder(next) && current.end() === next.start()) continue;
            if(placeholder(current) && parens(next) && current.end() === next.child('open').start()) { parts.push(d.soft()); parts.push(d.text('')); continue; }
            if(cv === 'with' && parens(next)) { parts = [d.concat([d.fill(parts),d.text(' ')])]; continue; }
            if(cv.endsWith('#') && nv === '{' && parens(next.child('group'))) continue;
            if(inlineValue(next) && nextNext.index < 0) continue;
            if(at.index < 0 && current.type() === 'value-comment' && !current.truth('inline') && groups.slice(0,i).every(g => new Node(this.tree,g).type() === 'value-comment')) { parts.push(d.dedent(d.line())); parts.push(d.text('')); continue; }
            parts.push(d.line()); parts.push(d.text(''));
        }
        if(groups.some(g => inlineValue(new Node(this.tree,g)))) append(d.add('break'));
        if(didBreak) parts = [d.text(''),d.hard(),...parts];
        if(directive) return d.group(d.indent(d.concat(parts)));
        if(groups.length === 2 && at.string('name') === 'import' && atIndex(0).string('value') === 'url') return d.group(d.fill(parts));
        return d.group(d.indent(d.fill(parts)));
    }
}
export type FormatResult = { readonly kind: 'Formatted'; readonly text: string } | { readonly kind: 'Refused'; readonly message: string };
export function format(text: string, scss: boolean = false, options: PrintOptions = defaults): FormatResult {
    const parsed = compose(text, scss);
    if(parsed.kind === 'Refused') return parsed;
    const front = new Node(parsed.tree,parsed.tree.root).child('frontMatter');
    if(front.string('language') === 'yaml' && front.string('value').trim() !== '') return {kind:'Refused',message:quote((scss ? 'scss' : 'css') + ': yaml front matter is formatted by the yaml printer, which css.Format cannot reach')};
    try {
        const printer = new Printer(parsed.tree,text,options,scss); const root = new Path(parsed.tree,[parsed.tree.root],[''],[-1]);
        return {kind:'Formatted',text:printer.docs.print(printer.print(root),options.width,options.tabWidth,options.tabs)};
    } catch(error) { return {kind:'Refused',message:quote((scss ? 'scss' : 'css') + ': ' + (error instanceof Error ? error.message : 'printer error'))}; }
}
