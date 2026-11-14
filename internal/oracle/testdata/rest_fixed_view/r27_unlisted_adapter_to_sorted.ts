console.log([5, 3, 9].toSorted((left: number, ...rest: number[]) => left - (rest[0] ?? 0)).join(","));
