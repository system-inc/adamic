package fuzz

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// GenerateMoves isolates the move vocabulary for a bounded seed campaign. The
// normal generator also selects it when both parallel and moves are enabled.
func GenerateMoves(seed uint64) *Program {
	g := &generator{random: rand.New(rand.NewPCG(seed, 0x61646d6963)), without: map[string]bool{}}
	return g.movesProgram()
}

const movesRefusePrefix = "// moves-refuse:"
const movesFixPrefix = "// moves-fix:"

// moveRefusal keeps the complete expected diagnostic in source, so saved and
// shrunk programs retain their oracle even when run against another checkout.
func moveRefusal(source string) (what, fix string) {
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, movesRefusePrefix) {
			what = strings.TrimSpace(strings.TrimPrefix(line, movesRefusePrefix))
		}
		if strings.HasPrefix(line, movesFixPrefix) {
			fix = strings.TrimSpace(strings.TrimPrefix(line, movesFixPrefix))
		}
	}
	return
}

// Exact comparison includes the path and fix, not just a diagnostic substring.
func judgeMoveRefusal(what, fix string, lowered Run) Outcome {
	refusal := compilerRefusal(lowered)
	if lowered.TimedOut || refusal.Verdict == Finding {
		return refusal
	}
	expected := what + "; " + fix
	_, actual, present := strings.Cut(strings.TrimSpace(string(lowered.Stderr)), "Adamic 0.1 refuses ")
	if !present || actual != expected {
		return Outcome{Verdict: Finding, Key: "move refusal differs", Detail: "expected " + expected + "\n" + firstLines(string(lowered.Stderr), 8)}
	}
	return Outcome{Verdict: Refused, Key: expected, Detail: firstLines(string(lowered.Stderr), 4)}
}

func (g *generator) movesProgram() *Program {
	kind := g.pick("accepted", "accepted", "accepted", "accepted", "after", "alias", "global", "closure", "nested-object", "nested-array", "nested-map")
	items, results, run := g.name("items"), g.name("results"), g.name("runMoves")
	value, flag, child := g.name("value"), g.name("done"), g.name("child")
	element, observer := g.name("element"), g.name("observer")
	count := 256 + g.random.IntN(65) // Several runtime grains, even for refused race witnesses.
	multiplier, rounds := 3+g.random.IntN(15), 1000+g.random.IntN(1001)
	program := &Program{Feature: "moves"}
	add := func(s *Statement) { program.Block.Statements = append(program.Block.Statements, s) }
	add(statement("import { parallelMap } from 'adamic';"))
	add(statement("// moves-shape: " + kind))
	body := &Block{}
	if kind == "global" {
		add(statement(fmt.Sprintf("const %s = { %s: %d, %s: false };", element, value, g.random.IntN(100), flag)))
	} else if kind == "alias" || kind == "closure" || kind == "nested-object" {
		body.Statements = append(body.Statements, statement(fmt.Sprintf("const %s = { %s: %d, %s: false };", element, value, g.random.IntN(100), flag)))
	}
	if kind == "alias" {
		body.Statements = append(body.Statements, statement("const "+observer+" = "+element+";"))
	}
	if kind == "closure" {
		body.Statements = append(body.Statements, statement("const "+observer+" = () => "+element+"."+value+";"))
	}
	entries := make([]string, count)
	for index := range entries {
		record := fmt.Sprintf("{ %s: %d, %s: %s }", value, g.random.IntN(1000), flag, g.pick("true", "false"))
		switch kind {
		case "alias", "global", "closure":
			entries[index] = element
		case "nested-object":
			entries[index] = fmt.Sprintf("{ %s: %s, %s: %d }", child, element, value, index)
		case "nested-array":
			entries[index] = "[" + record + "]"
		case "nested-map":
			entries[index] = fmt.Sprintf("new Map<number, { %s: number; %s: boolean }>([[%d, %s]])", value, flag, index, record)
		default:
			entries[index] = record
		}
	}
	body.Statements = append(body.Statements, statement("const "+items+" = [\n\t\t"+strings.Join(entries, ",\n\t\t")+"\n\t];"))
	item, index, step := g.name("item"), g.name("index"), g.name("step")
	work := &Block{}
	path := item + "." + value
	if kind == "nested-object" {
		path = item + "." + child + "." + value
	}
	if kind != "nested-array" && kind != "nested-map" {
		work.Statements = append(work.Statements, statement(fmt.Sprintf("for (let %s = 0; %s < %d; %s++) @b", step, step, rounds, step), &Block{Statements: []*Statement{
			statement(fmt.Sprintf("%s = (%s * %d + %s + %s) %% 65521;", path, path, multiplier, step, index)),
		}}))
		if kind != "nested-object" {
			work.Statements = append(work.Statements, statement(item+"."+flag+" = !"+item+"."+flag+";"))
		}
	}
	work.Statements = append(work.Statements, statement("return "+item+";"))
	body.Statements = append(body.Statements, statement("const "+results+" = parallelMap("+items+", ("+item+", "+index+") => @b);", work))
	switch kind {
	case "after":
		program.Refusal = "use after move: " + items + ".length; " + items + " was moved into parallelMap"
		program.RefusalFix = "don't use it after the parallelMap"
		body.Statements = append(body.Statements, statement("console.log(`${"+items+".length}`);"))
	case "alias", "global", "closure":
		program.Refusal = "cannot move " + items + "[0]: element is not a fresh object literal with only scalar literal fields"
	case "nested-object":
		program.Refusal = "cannot move " + items + "[0]." + child + ": nested reachable ownership is not proven"
	case "nested-array":
		program.Refusal = "cannot move " + items + "[0][]: nested array ownership is not proven"
	case "nested-map":
		program.Refusal = "cannot move " + items + "[0].values[]: nested Map ownership is not proven"
	}
	if program.Refusal != "" {
		if program.RefusalFix == "" {
			program.RefusalFix = "return it through the results"
		}
		add(statement(movesRefusePrefix + " " + program.Refusal))
		add(statement(movesFixPrefix + " " + program.RefusalFix))
		body.Statements = append(body.Statements, statement("console.log(`${"+results+".length}`);"))
		if kind == "alias" {
			body.Statements = append(body.Statements, statement("console.log(`${"+observer+"."+value+"}`);"))
		}
		if kind == "closure" {
			body.Statements = append(body.Statements, statement("console.log(`${"+observer+"()}`);"))
		}
		if kind == "global" || kind == "nested-object" {
			body.Statements = append(body.Statements, statement("console.log(`${"+element+"."+value+"}`);"))
		}
	} else {
		digest, at, row := g.name("digest"), g.name("at"), g.name("row")
		body.Statements = append(body.Statements, statement("let "+digest+" = 0;"), statement("for (let "+at+" = 0; "+at+" < "+results+".length; "+at+"++) @b", &Block{Statements: []*Statement{
			statement("const " + row + " = " + results + "[" + at + "];"),
			statement("if ("+row+" !== undefined) @b", &Block{Statements: []*Statement{
				statement(digest + " = (" + digest + " * 31 + " + row + "." + value + " + (" + row + "." + flag + " ? 1 : 0)) % 1000000007;"),
			}}),
		}}), statement("console.log(`${"+results+".length}:${"+digest+"}:${"+results+"[0]?."+value+"}:${"+results+"["+results+".length - 1]?."+value+"}`);"))
	}
	add(statement("function "+run+"(): void @b", body))
	add(statement(run + "();"))
	return program
}
