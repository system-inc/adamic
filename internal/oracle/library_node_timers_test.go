package oracle

// timerSourceFixtures are setTimeout and clearTimeout called from source (#995spev), held to Node by
// the ordinary families: native under the sanitizers and the leak check, the JavaScript backend,
// release, gob and counted. Each schedules a timer, so wasm32-wasi refuses it at setTimeout.
var timerSourceFixtures = []string{"forwarded", "fewer_parameters", "clear_then_set", "order", "callback_throw", "clear_kinds"}

func init() {
	for _, name := range timerSourceFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/timers_source/" + name + ".a", true, false})
	}
}
