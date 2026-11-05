const v: readonly (number | string)[] = Object.values({ a: 1 } as {});
for (const x of v) { console.log(String(x)); }
