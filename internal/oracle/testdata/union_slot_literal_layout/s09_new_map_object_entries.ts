enum Kind { A, B }
const o: { type: Map<string, number | string> } = { type: new Map(Object.entries({ a: Kind.A, b: Kind.B })) };
for (const [key, value] of o.type) { console.log(key + " = " + String(value)); }
