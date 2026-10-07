package fuzz

// Each block is part of the shrinker's tree. The cadence covers the Cartesian product of
// effects and argument shapes in 50 seeds; random labels and suffixes vary the actual values.
// Both implementations and a super call run, so a static signature cannot hide an escape.
func (g *generator) overridesProgram() []*Statement {
	fresh := func(mark string) string { return g.ownFreshBox(mark) }
	parameters := "(box: OverrideBox, items: OverrideBox[], callback: () => void): OverrideBox "
	read := blockOf(statement("return { v: `${box.v}:${items.length}` };"))
	body := &Block{}
	add := func(s string) { body.Statements = append(body.Statements, statement(s)) }
	switch g.seed % 10 {
	case 0:
		add("items[0] = " + fresh("write") + ";")
		add("box.v = `${box.v}" + g.pick("!", "?", "+") + "`;")
	case 1:
		add("this.saved = box;")
	case 2:
		add("overrideGlobal = box;")
	case 3:
		add("overrideArray.push(box);")
	case 4:
		add("overrideMap.set('kept', box);")
	case 5:
		add("overrideGlobal = box;")
		add("throw new Error(box.v);")
	case 6:
		add("overrideClosures.push(() => box.v);")
	case 7:
		add("callback();")
	case 8:
		add("return box;")
	case 9:
		// The override consumes its array parameter, although the base only reads it.
		add("const mapped = items.map((item) => ({ v: `${item.v}!` }));")
		add("return mapped[0] ?? " + fresh("empty") + ";")
	}
	if g.seed%10 != 5 && g.seed%10 != 8 && g.seed%10 != 9 {
		add("return { v: `${box.v}:${items.length}` };")
	}
	call := "worker.run(@e, items, callback)"
	argument := text(Other, "borrowed")
	switch (g.seed / 10) % 5 {
	case 1:
		argument = text(Other, "items[0] ?? "+fresh("missing"))
	case 2:
		argument = text(Other, "overrideNarrowed")
	case 3:
		argument = text(Other, fresh("literal"))
	case 4:
		argument = text(Other, "{ ...borrowed, v: `${borrowed.v}+` }")
	}
	callbackBody := &Block{}
	if g.seed%10 == 7 {
		callbackBody.Statements = append(callbackBody.Statements, statement("items[0] = "+fresh("callback")+";"))
	}
	exercise := blockOf(
		statement("const items: OverrideBox[] = ["+fresh("first")+", "+fresh("second")+"];"),
		statement("const borrowed = items[0] ?? "+fresh("fallback")+";"),
		statement("const callback = (): void => @b;", callbackBody),
		statement("if (overrideNarrowed !== undefined) @b", blockOf(
			statement("try @b catch (error) @b", blockOf(
				statement("const result = "+call+";", argument),
				statement("console.log(result.v);"),
			), blockOf(statement("console.log('override caught');"))),
		)),
		statement("console.log(borrowed.v);"),
	)
	return []*Statement{
		statement("interface OverrideBox { v: string; }"),
		statement("interface OverrideWorker { run" + parameters + "; }"),
		statement("let overrideGlobal: OverrideBox = " + fresh("global") + ";"),
		statement("let overrideNarrowed: OverrideBox | undefined = " + fresh("narrowed") + ";"),
		statement("const overrideArray: OverrideBox[] = [];"),
		statement("const overrideMap = new Map<string, OverrideBox>();"),
		statement("const overrideClosures: (() => string)[] = [];"),
		statement("class OverrideBase implements OverrideWorker @b", blockOf(
			statement("saved: OverrideBox = "+fresh("field")+";"),
			statement("run"+parameters+"@b", read),
		)),
		statement("class OverrideChild extends OverrideBase @b", blockOf(
			statement("run"+parameters+"@b", body),
			statement("throughSuper(): void @b", blockOf(
				statement("const items: OverrideBox[] = ["+fresh("super")+"];"),
				statement("const borrowed = items[0] ?? "+fresh("fallback")+";"),
				statement("console.log(super.run(borrowed, items, () => {}).v);"),
				statement("console.log(borrowed.v);"),
			)),
		)),
		statement("class OverrideReader implements OverrideWorker @b", blockOf(statement("run"+parameters+"@b", read.clone()))),
		statement("function overrideViaBase(worker: OverrideBase): void @b", exercise),
		statement("function overrideViaInterface(worker: OverrideWorker): void @b", exercise.clone()),
		statement("const overrideChild = new OverrideChild();"),
		statement("overrideViaBase(new OverrideBase());"),
		statement("overrideViaBase(overrideChild);"),
		// Single-slot destinations are observed before the next call can overwrite them.
		statement("console.log(overrideChild.saved.v);"),
		statement("console.log(overrideGlobal.v);"),
		statement("for (const [key, kept] of overrideMap) @b", blockOf(statement("console.log(`${key}:${kept.v}`);"))),
		statement("overrideViaInterface(new OverrideReader());"),
		statement("overrideViaInterface(overrideChild);"),
		statement("overrideChild.throughSuper();"),
		// The caller's arrays and temporaries have gone by now. All escape destinations are read.
		statement("console.log(overrideChild.saved.v);"),
		statement("console.log(overrideGlobal.v);"),
		statement("for (const kept of overrideArray) @b", blockOf(statement("console.log(kept.v);"))),
		statement("for (const [key, kept] of overrideMap) @b", blockOf(statement("console.log(`${key}:${kept.v}`);"))),
		statement("for (const later of overrideClosures) @b", blockOf(statement("console.log(later());"))),
	}
}
