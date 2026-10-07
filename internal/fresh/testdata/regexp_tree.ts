interface Tree {
    readonly label: string;
    readonly children: Tree[];
}
function build(text: string, depth: number): Tree {
    const tree: Tree = { label: text.replace(/\d+/g, '#'), children: [] };
    if (depth > 0 && /\d/.test(text)) {
        tree.children.push(build(`${text}${depth}`, depth - 1));
    }
    return tree;
}
const root = build(`leaf${3}`, 3);
console.log(root.label);
console.log(`${root.children.length}`);
const match = /(?<word>leaf)(\d+)/d.exec(`leaf${3}`);
if (match !== null) {
    console.log(match[0]);
    console.log(`${match.index}`);
    console.log(match.groups?.word ?? 'missing');
    console.log(`${match.indices?.length ?? 0}`);
}
console.log(`${'a1b2'.match(/\d/g)?.length ?? 0}`);
console.log(`${'a1b2'.split(/\d/).length}`);
console.log(`${'a1'.search(/\d/)}`);
console.log('a1b2'.replaceAll(/\d/g, '#'));
const iterator = 'a1b2'.matchAll(/\d/g);
const first = iterator.next();
console.log(`${first.done === true}`);
console.log(first.value === undefined ? 'missing' : first.value[0] ?? 'missing');
console.log(`${iterator.next().done === true}`);
console.log(`${iterator.next().done === true}`);
