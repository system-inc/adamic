// An object pattern over a Set is no plain array road, so it keeps its stop by name (#rm22wzv).
const seen = new Set<{ a: number }>([{ a: 1 }]);
for (const { a } of seen) {
    console.log(String(a));
}
