// The construction-time fields of one mdast node.
import type { ParsedFrontMatter } from './parseFrontMatter.ts';
import { copyTokenPoint, tokenPoint, type TokenPointInterface } from './tokenArena.ts';
export class MdastNode {
    type: string;
    readonly children: number[] = [];
    parent = false;
    literal = false;
    valueNull = false;
    value = '';
    depth = 0;
    ordered = false;
    startValue = 0;
    hasStartValue = false;
    spread = false;
    checked = false;
    hasChecked = false;
    lang = '';
    hasLang = false;
    meta = '';
    hasMeta = false;
    url = '';
    title = '';
    hasTitle = false;
    alt = '';
    hasAlt = false;
    label = '';
    hasLabel = false;
    identifier = '';
    referenceType = '';
    align: string[] = [];
    frontMatter: ParsedFrontMatter | undefined = undefined;
    start = tokenPoint(0, 0, 0, 0, 0);
    end = tokenPoint(0, 0, 0, 0, 0);
    constructor(type: string) {
        this.type = type;
    }
    setStart(point: TokenPointInterface): void {
        this.start = copyTokenPoint(point);
    }
    setEnd(point: TokenPointInterface): void {
        this.end = copyTokenPoint(point);
    }
}
