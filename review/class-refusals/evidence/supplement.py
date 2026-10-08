from pathlib import Path
import json
b=Path('/tmp/class-refusals'); p=b/'probes'; cases=json.loads((p/'cases.json').read_text())
def add(name,kind,source,rule=''):
 path=p/(name+'.a'); path.write_text(source+'\n'); cases.append(dict(name=name,kind=kind,path=str(path),rule=rule))
add('super_unused_arrow','suspect','class Base { describe(): string { const later = () => this.score(); return "safe"; } score(): number { return 0; } } class Derived extends Base { snap = super.describe(); count = 5; score(): number { return this.count + 1; } } console.log(new Derived().snap);')
add('super_dead_branch','suspect','class Base { describe(): string { if (false) return `score ${this.score()}`; return "safe"; } score(): number { return 0; } } class Derived extends Base { snap = super.describe(); count = 5; score(): number { return this.count + 1; } } console.log(new Derived().snap);')
source=(p/'keys_interface.a').read_text().split('interface View')[0]
add('keys_optional_call','refuse',source+'const view: { readonly value: number } = source; console.log(Object.keys?.(view).join(","));','symbol-key-view')
add('keys_nested_arrow','refuse',source+'function outer(v: { readonly value: number }): string { const inner = (w: { readonly value: number }) => Object.keys(w).join(","); return inner(v); } console.log(outer(source));','symbol-key-view')
box='class Box { #secret: string; constructor(secret: string) { this.#secret = secret; } '
add('private_optional_holder','refuse',box+'static peek(holder: { readonly value: Box } | undefined): string { const box = holder?.value; if (box === undefined) return "none"; return box.#secret; } } console.log(Box.peek({ value: new Box(`s${1}`) }));','private-instance-from-static')
add('private_interface','refuse','interface Holder { readonly value: Box; } '+box+'static peek(holder: Holder): string { return holder.value.#secret; } } console.log(Box.peek({ value: new Box(`s${1}`) }));','private-instance-from-static')
it=(p/'iterator_optional.a').read_text().split('class Holder')[0]
add('iterator_optional_chain','refuse',it+'class Holder { value: BaseIterator = new ScaledIterator(); } function use(holder: Holder | undefined): void { console.log([...(holder?.value ?? new BaseIterator())].join(",")); } use(new Holder());','iterator-receiver-origin')
it=Path('internal/oracle/testdata/iterators_hidden_return.a').read_text().replace('return this;', 'return this.index >= 0 ? this : this;')
add('iterator_conditional_close','suspect',it)
(p/'cases.json').write_text(json.dumps(cases,indent=2)); (b/'supplement-cases.json').write_text(json.dumps(cases[-8:],indent=2))
