const map = new Map<string | number, boolean>();
map.set(1, true);
console.log(String(map.has(1)));
console.log(String(map.has("1")));
console.log(String(map.get(1)));
console.log(String(map.get("1")));
console.log(String(map.delete("1")));
