interface Required { count: number }
interface Optional { count: number | undefined }
const narrow: Required[] = [{ count: 1 }];
function store(values: Optional[], value: number | undefined): string { const item = values[0]; if (item) { item.count = value; return (item.count ?? 0).toString(); } return ""; }
console.log(store(narrow, 2));
