interface Named { readonly name: string }
const holder: Named[] = [];
const made = { name: "v", back: holder };
holder.push(made);
for (const item of holder) {
    console.log(`${holder.length} ${item.name}`);
}
