interface Named { readonly name: string }
const holder: Named[] = [];
const extra = { back: holder };
holder.push({ ...extra, name: "c" });
console.log(String(holder.length), holder[0].name);
