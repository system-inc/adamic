const proto = { inherited: 1 };
const odd = { "__proto__": proto, own: 2 };
console.log(Object.keys(odd).join(",") + " " + String(Object.values(odd).length));
