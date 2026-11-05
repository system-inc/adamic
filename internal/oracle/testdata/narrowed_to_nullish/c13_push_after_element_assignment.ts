function describe(content: string | null | undefined): string { return content === null ? 'null' : content === undefined ? 'undefined' : content; }
const entries: (string | null)[] = ['a', null, ''];
entries[0] = null;
entries.push('tail');
console.log(`${describe(entries[0])}|${describe(entries[3])}`);
