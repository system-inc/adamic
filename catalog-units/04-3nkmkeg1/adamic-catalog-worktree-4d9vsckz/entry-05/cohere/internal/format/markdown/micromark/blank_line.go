package micromark

// blankLine is micromark-core-commonmark/lib/blank-line.js.
var blankLine = &Construct{Partial: true, Tokenize: tokenizeBlankLine}

// tokenizeBlankLine is tried at the start of every line of flow, so its state is a blankLineRun from the
// parse's Memory, whose states were bound when the slot was made, and a try allocates nothing (#vbjv3d6).
func tokenizeBlankLine(self *Self, effects *Effects, ok State, nok State) State {
	run := effects.memory().blankLine()
	run.effects, run.ok, run.nok = effects, ok, nok
	return run.startState
}

// blankLineRun is one blank line: upstream's closure state, with its states made once per slot.
type blankLineRun struct {
	effects *Effects
	ok, nok State

	startState, afterState State
}

func newBlankLineRun() *blankLineRun {
	run := &blankLineRun{}
	run.startState, run.afterState = run.start, run.after
	return run
}

// release zeroes the run for its next parse, keeping its bound states.
func (run *blankLineRun) release() {
	*run = blankLineRun{startState: run.startState, afterState: run.afterState}
}

func (run *blankLineRun) start(code Code) State {
	if markdownSpace(code) {
		return factorySpace(run.effects, run.afterState, TypeLinePrefix, 0)(code)
	}
	return run.after(code)
}

func (run *blankLineRun) after(code Code) State {
	if code == CodeEof || markdownLineEnding(code) {
		return run.ok(code)
	}
	return run.nok(code)
}
