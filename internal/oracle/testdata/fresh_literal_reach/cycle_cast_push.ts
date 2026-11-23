interface Named { readonly name: string }
const holder: Named[] = [];
holder.push({ name: "c", back: holder } as Named);
for (const item of holder) {
    console.log(`${holder.length} ${item.name}`);
}
