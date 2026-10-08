const groups = Object.groupBy([1, 2], (value: number): string => value === 1 ? 'one' : 'other');
