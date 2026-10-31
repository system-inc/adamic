interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class SubPoint extends Point { describe(): string { return "subpoint"; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
places[key] = new SubPoint();
console.log(places.names.describe().toUpperCase());
