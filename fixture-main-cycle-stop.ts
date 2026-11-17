// @ts-check
interface Node { rs: Node[] | Record<string, number>[] }
const a: Node = { rs: [] };
a.rs = [a];
const record: Record<string, number> = { k: 1 };
a.rs = [record];
