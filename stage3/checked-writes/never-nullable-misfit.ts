const empty: never[] = [];
const view: { name: string }[] | undefined = empty;
view.push({name: "new" + "new"});
console.log(`${view.length}`);
