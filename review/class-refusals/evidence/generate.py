from pathlib import Path
import json
root=Path('/workspace/adamic'); out=Path('/tmp/class-refusals/probes'); out.mkdir(exist_ok=True)
cases=[]
def add(name,kind,source,rule=''):
 p=out/(name+'.a'); p.write_text(source+'\n'); cases.append(dict(name=name,kind=kind,path=str(p),rule=rule))
superbase='class Base { describe(): string { return `score ${this.score()}`; } score(): number { return 0; } }\n'
def sup(name,expr='super.describe()',method='score(): number { return this.count + 1; }',extra='',fields='snap = EXPR; count: number = 5;'):
 add('super_'+name,'refuse',extra+superbase+'class Derived extends Base { '+fields.replace('EXPR',expr)+' '+method+' }\nconsole.log(new Derived().snap);','uninitialized field')
sup('parentheses','(super.describe())')
sup('optional','super.describe?.()')
sup('arrow_call','(() => super.describe())()')
sup('getter_override',method='get number(): number { return this.count + 1; } score(): number { return this.number; }')
sup('nested',method='score(): number { return this.middle(); } middle(): number { return this.last; } get last(): number { return this.count + 1; }')
sup('generic',method='score(): number { return identity(this.count) + 1; }',extra='function identity<T>(value: T): T { return value; }\n')
sup('interface',method='score(): number { const view: Count = { count: this.count }; return view.count + 1; }',extra='interface Count { readonly count: number; }\n')
sup('union',method='score(): number { const value: number | undefined = this.count; return (value ?? 9) + 1; }')
add('super_no_override','safe',superbase+'class Derived extends Base { snap = super.describe(); count = 5; } console.log(new Derived().snap);')
add('super_ordered','safe',superbase+'class Derived extends Base { count = 5; snap = super.describe(); score(): number { return this.count + 1; } } console.log(new Derived().snap);')
add('super_ordered_getter','safe','class Base { get summary(): string { return `score ${this.score()}`; } score(): number { return 0; } } class Derived extends Base { count = 5; snap = super.summary; score(): number { return this.count + 1; } } console.log(new Derived().snap);')
add('super_deferred','suspect',superbase+'class Derived extends Base { snap = () => super.describe(); count = 5; score(): number { return this.count + 1; } } console.log(new Derived().snap());')
add('super_ignored_argument','suspect','class Base { describe(value: number): string { return `value ${value}`; } } class Derived extends Base { snap = super.describe(this.count); count = 5; } console.log(new Derived().snap);')
add('super_inherited_reset','suspect','class Base { count = 4; describe(): string { return `score ${this.score()}`; } score(): number { return this.count; } } class Derived extends Base { snap = super.describe(); count = 5; score(): number { return this.count + 1; } } console.log(new Derived().snap);')
it=(root/'internal/oracle/testdata/iterators_override_this.a').read_text().split('for(const value')[0]
uses={'for_of':'for (const v of new ScaledIterator()) { console.log(`${v}`); break; }','spread':'console.log([...new ScaledIterator()].join(","));','from':'console.log(Array.from(new ScaledIterator()).join(","));','destructure':'const [v = -1] = new ScaledIterator(); console.log(`${v}`);','alias':'const a = new ScaledIterator(); const b = a; console.log([...b].join(","));','arrow':'const make = () => new ScaledIterator(); console.log([...make()].join(","));','getter':'class Holder { get value(): BaseIterator { return new ScaledIterator(); } } console.log([...new Holder().value].join(","));','generic':'function erase<T extends BaseIterator>(v: T): BaseIterator { return v; } console.log([...erase(new ScaledIterator())].join(","));','interface':'interface View { [Symbol.iterator](): BaseIterator; } function use(v: View): void { console.log([...v].join(",")); } use(new ScaledIterator());','union':'function choose(flag: boolean): BaseIterator | ScaledIterator { return flag ? new ScaledIterator() : new BaseIterator(); } console.log([...choose(true)].join(","));','optional':'class Holder { value: BaseIterator | undefined = new ScaledIterator(); } console.log(Array.from(new Holder().value ?? new BaseIterator()).join(","));'}
for name,use in uses.items(): add('iterator_'+name,'refuse',it+use,'iterator-receiver-origin')
add('iterator_return_alias','refuse',it.replace('return this;', 'const receiver = this; return (receiver);')+uses['spread'],'iterator-receiver-origin')
add('iterator_return_arrow','suspect',it.replace('return this;', 'const receiver = () => this; return receiver();')+uses['spread'])
add('iterator_return_conditional','suspect',it.replace('return this;', 'return this.index >= 0 ? this : this;')+uses['spread'])
for name,use in {'base':'console.log([...new BaseIterator()].join(","));','unchanged':'class Unchanged extends BaseIterator {} console.log([...new Unchanged()].join(","));'}.items(): add('iterator_'+name,'safe',it+use)
add('iterator_closed_parameter','safe',it.split('class ScaledIterator')[0]+'function use(v: BaseIterator): void { console.log([...v].join(",")); } use(new BaseIterator());')
add('iterator_base_parameter','suspect',it+'function use(v: BaseIterator): void { console.log([...v].join(",")); } use(new BaseIterator());')
add('iterator_factory','safe',it.replace('return this;', 'return new BaseIterator();')+uses['spread'])
source='const source = { value: 1, [Symbol.iterator]() { let index = 0; return { next() { index++; return { value: index, done: index > 2 }; } }; } };\n'
keyuses={'interface':'interface View { readonly value: number; } function keys(v: View): string { return Object.keys(v).join(","); } console.log(keys(source));','arrow':'const keys = (v: { readonly value: number }) => Object.keys(v).join(","); console.log(keys(source));','generic':'function keys<T extends { readonly value: number }>(v: T): string { return Object.keys(v).join(","); } const view: { readonly value: number } = source; console.log(keys(view));','getter':'class Holder { get view(): { readonly value: number } { return source; } } console.log(Object.keys(new Holder().view).join(","));','union':'function keys(v: { readonly value: number } | { readonly other: number }): string { return Object.keys(v).join(","); } console.log(keys(source));','optional':'class Holder { view: { readonly value: number } | undefined = source; } const view = new Holder().view; if (view !== undefined) console.log(Object.keys(view).join(","));','nested':'function outer(v: { readonly value: number }): string { function inner(w: { readonly value: number }): string { return Object.keys(w).join(","); } return inner(v); } console.log(outer(source));'}
for name,use in keyuses.items(): add('keys_'+name,'refuse',source+use,'symbol-key-view')
add('keys_plain','safe','const plain = { value: 1 }; console.log(Object.keys(plain).join(","));')
add('keys_copy','safe',source+'console.log(Object.keys({ value: source.value }).join(","));')
add('keys_unrelated','suspect',source+'const plain = { value: 7 }; console.log(Object.keys(plain).join(",")); console.log([...source].join(","));')
add('keys_class','safe','class Ranged { readonly value = 3; [Symbol.iterator](): Ranged { return this; } next(): { readonly value: number; readonly done: boolean } { return { value: 0, done: true }; } } console.log(Object.keys(new Ranged()).join(","));')
private='class Box { #secret: string; constructor(secret: string) { this.#secret = secret; } '
for name,member,call in [('arrow','static peek(box: Box): string { const read = () => box.#secret; return read(); }','Box.peek(new Box(`s${1}`))'),('getter','static get value(): string { return new Box(`s${1}`).#secret; }','Box.value'),('optional','static peek(box: Box | undefined): string { return box?.#secret ?? "none"; }','Box.peek(new Box(`s${1}`))'),('generic','static peek<T extends Box>(box: T): string { return box.#secret; }','Box.peek(new Box(`s${1}`))'),('union','static peek(box: Box | undefined): string { if (box === undefined) return "none"; return box.#secret; }','Box.peek(new Box(`s${1}`))'),('nested','static peek(box: Box): string { function read(value: Box): string { return value.#secret; } return read(box); }','Box.peek(new Box(`s${1}`))'),('write','static peek(box: Box): string { box.#secret = `t${2}`; return box.reveal(); } reveal(): string { return this.#secret; }','Box.peek(new Box(`s${1}`))')]: add('private_'+name,'refuse',private+member+' } console.log('+call+');','private-instance-from-static')
add('private_instance','safe',private+'reveal(): string { return this.#secret; } static peek(box: Box): string { return box.reveal(); } } console.log(Box.peek(new Box(`s${1}`)));')
add('private_static','safe','class Box { static #secret = `s${1}`; static peek(): string { return Box.#secret; } } console.log(Box.peek());')
(out/'cases.json').write_text(json.dumps(cases,indent=2)); print(len(cases),'cases')
