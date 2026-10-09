const empty: never[] = [];
const view: { name: string }[] | undefined = empty;
view.fill({name: "new" + "new"});
console.log(`${view.length}`);
