package oracle

// comparatorRestFixtures pin closureComparator's end of statement (#kwns2n5): a sort comparator that
// collects its second argument as a rest gets an adapter that packs a tail for every comparison, and
// the adapter lets go of it once the comparison returns. They're held to Node by the ordinary
// families, and native under the sanitizers and the leak check is the one that can fail on a kept
// tail. Mutant, run by hand: arguments_length.go's closureComparator without its adapter.end(), and
// LeakSanitizer reports every comparison's tail on rest_sort.a (the probe J-251pqkd found 10 objects
// on six numbers at main 1046dcfd).
var comparatorRestFixtures = []string{"rest_sort"}

func init() {
	for _, name := range comparatorRestFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/comparator_rest/" + name + ".a", true, false})
	}
}
