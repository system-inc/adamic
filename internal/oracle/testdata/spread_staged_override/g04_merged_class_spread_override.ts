// The publish gap through an interface merged into a class. The merge declares label, which the constructor
// never writes, so the class's layout reserves it as an absent slot (internal/lower/merged_fields.go:201).
// The spread copies the instance's public fields, label still absent, and the explicit label is written in.
class Box {
  kind = 1;
}
interface Box {
  label: string;
}
const b = new Box();
const m = { ...b, label: 'x' };
console.log(String(m.label)); // Node: x
console.log(Object.keys(m).join()); // Node: kind,label
console.log(String(JSON.stringify(m))); // Node: {"kind":1,"label":"x"}
console.log(String('label' in m)); // Node: true
