function copy(source: { readonly code: number } | undefined) {
    return { ...source };
}
const made = copy(undefined);
console.log(String("code" in made));
made.code = 8;
console.log(String("code" in made));
console.log(String(made.code));
console.log(JSON.stringify(copy(undefined)) + " " + JSON.stringify(made));
