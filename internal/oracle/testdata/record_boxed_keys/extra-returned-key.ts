interface Values { [name: string]: string }
const record: Values = { "2": "number" };
function key(): keyof Values { return 2; }
console.log(String(record[key()]));
