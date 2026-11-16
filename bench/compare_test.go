package main

import (
	"strings"
	"testing"
	"time"
)

// fakeSamples makes n samples of a program whose baseline takes 100 ms of user time and whose
// candidate takes ratios[i] times that, every run finishing well above the floor with Node's answer.
func fakeSamples(ratios []float64) [][3]measurement {
	samples := [][3]measurement{}
	for _, ratio := range ratios {
		baseline := 100 * time.Millisecond
		candidate := time.Duration(float64(baseline) * ratio)
		samples = append(samples, [3]measurement{
			{elapsed: baseline, user: baseline, stdout: "42\n"},
			{elapsed: candidate, user: candidate, stdout: "42\n"},
			{elapsed: 2 * baseline, user: 2 * baseline, stdout: "42\n"},
		})
	}
	return samples
}

var fakeFloors = [3]measurement{{elapsed: time.Millisecond}, {elapsed: time.Millisecond}, {elapsed: 20 * time.Millisecond}}

// k is the largest index with P(Binomial(n, 1/2) < k) <= 0.025, by hand:
// n = 6: P(B < 1) = 1/64 = 0.016, P(B < 2) = 7/64 = 0.109, so k = 1, the smallest and largest.
// n = 15: P(B < 4) = (1 + 15 + 105 + 455)/32768 = 576/32768 = 0.0176, and P(B < 5) adds
// C(15, 4) = 1365 for 1941/32768 = 0.0592, so k = 4: the 4th and 12th.
// n = 30: P(B < 10) = 22964087/2^30 = 0.0214, and P(B < 11) adds C(30, 10) = 30045015 for
// 53009102/2^30 = 0.0494, so k = 10: the 10th and 21st.
func TestSignTestRankIsTheBinomialTailsOrderStatistic(t *testing.T) {
	t.Parallel()
	for pairs, want := range map[int]int{6: 1, 15: 4, 30: 10} {
		rank, err := signTestRank(pairs)
		if err != nil {
			t.Fatalf("n = %d was refused: %v", pairs, err)
		}
		if rank != want {
			t.Fatalf("n = %d gave k = %d; want %d", pairs, rank, want)
		}
	}
}

// Below six pairs even the smallest and largest ratio leave more than 2.5% in a tail (1/32 at five),
// so there's no 95% interval, and the comparison says so instead of giving a verdict.
func TestRefusesFewerThanSixSamples(t *testing.T) {
	t.Parallel()
	for _, pairs := range []int{0, 1, 5} {
		if _, err := signTestRank(pairs); err == nil {
			t.Fatalf("n = %d gave a rank; want it refused", pairs)
		}
	}
	result := compareSamples(fakeSamples([]float64{1.5, 1.5, 1.5, 1.5, 1.5}), "user", fakeFloors, time.Microsecond)
	if result.verdict != "" || !strings.Contains(result.failure, "at least 6") {
		t.Fatalf("five samples gave verdict %q, failure %q; want them refused", result.verdict, result.failure)
	}
}

// A verdict needs the whole interval on one side of 1.00, and anything else reads as no change.
func TestVerdictNeedsTheWholeIntervalOnOneSide(t *testing.T) {
	t.Parallel()
	constant := func(ratio float64) []float64 {
		ratios := make([]float64, 15)
		for index := range ratios {
			ratios[index] = ratio
		}
		return ratios
	}
	cases := []struct {
		name   string
		ratios []float64
		want   string
	}{
		{"ratios near 1", []float64{0.99, 1.01, 1.00, 0.98, 1.02, 1.00, 0.995, 1.005, 0.99, 1.01, 1.00, 0.985, 1.015, 1.00, 1.00}, noChange},
		{"all 1.10", constant(1.10), "slower"},
		{"all 0.90", constant(0.90), "faster"},
		// Sorted: 0.90 0.92 0.93 0.95 0.96 0.97 0.98 1.00 1.02 1.03 1.04 1.05 1.08 1.10 1.12, so the 4th
		// is 0.95 and the 12th 1.05: the interval straddles 1.
		{"a mix straddling 1", []float64{1.10, 0.90, 1.02, 0.95, 1.05, 0.97, 1.12, 0.93, 1.00, 1.08, 0.96, 1.03, 0.92, 1.04, 0.98}, noChange},
		// Mostly slower, but four of fifteen below 1 put the 4th smallest under it.
		{"slower but for four", []float64{0.99, 0.98, 0.97, 0.96, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2}, noChange},
		// Three of fifteen below 1 leave the 4th smallest above it.
		{"slower but for three", []float64{0.99, 0.98, 0.97, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2, 1.2}, "slower"},
	}
	for _, test := range cases {
		result := compareSamples(fakeSamples(test.ratios), "user", fakeFloors, time.Microsecond)
		if result.failure != "" {
			t.Fatalf("%s: refused: %s", test.name, result.failure)
		}
		if result.verdict != test.want {
			t.Fatalf("%s: median %.3f, interval %.3f to %.3f read %q; want %q", test.name, result.median, result.low, result.high, result.verdict, test.want)
		}
	}
	median, low, high, err := signTestInterval([]float64{1.10, 0.90, 1.02, 0.95, 1.05, 0.97, 1.12, 0.93, 1.00, 1.08, 0.96, 1.03, 0.92, 1.04, 0.98})
	if err != nil || median != 1.00 || low != 0.95 || high != 1.05 {
		t.Fatalf("the straddling mix gave median %v, interval %v to %v, %v; want 1.00, 0.95 to 1.05", median, low, high, err)
	}
}

// A program whose answer differs from Node's, at either commit, in any sample, gets no verdict
// however clear its times are.
func TestAProgramThatDiffersFromNodeGetsNoVerdict(t *testing.T) {
	t.Parallel()
	for _, runner := range []int{baselineRunner, candidateRunner} {
		samples := fakeSamples([]float64{1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5})
		samples[3][runner].stdout = "41\n"
		result := compareSamples(samples, "user", fakeFloors, time.Microsecond)
		if result.verdict != "" || !strings.HasPrefix(result.failure, "differs from Node") {
			t.Fatalf("a %s that printed another answer than Node's gave verdict %q, failure %q; want differs from Node", runnerNames[runner], result.verdict, result.failure)
		}
	}
}

// run.go's refusals hold here: a reading within its runtime's floor measured nothing, so its
// program gets no verdict, never a ratio.
func TestRefusesAComparisonWithAReadingWithinItsFloor(t *testing.T) {
	t.Parallel()
	samples := fakeSamples([]float64{1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5, 1.5})
	samples[2][candidateRunner].elapsed = fakeFloors[candidateRunner].elapsed
	result := compareSamples(samples, "user", fakeFloors, time.Microsecond)
	if result.verdict != "" || !strings.Contains(result.failure, "measured nothing: ") {
		t.Fatalf("a candidate reading at its floor gave verdict %q, failure %q; want it refused", result.verdict, result.failure)
	}
}

// A program judged on wall reads its ratio from wall time, so a parallel program whose candidate
// spends more user time for less wall reads faster.
func TestAWallJudgedProgramReadsWallTime(t *testing.T) {
	t.Parallel()
	samples := fakeSamples([]float64{0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8, 0.8})
	for index := range samples {
		samples[index][candidateRunner].user = 4 * samples[index][baselineRunner].user
	}
	if result := compareSamples(samples, "wall", fakeFloors, time.Microsecond); result.verdict != "faster" {
		t.Fatalf("on wall, a candidate at 0.8 of the baseline's wall read %q, failure %q; want faster", result.verdict, result.failure)
	}
	if result := compareSamples(samples, "user", fakeFloors, time.Microsecond); result.verdict != "slower" {
		t.Fatalf("on user, a candidate at 4 times the baseline's user read %q, failure %q; want slower", result.verdict, result.failure)
	}
}

// A "no change" is printed with the change the run couldn't rule out, its interval's farther bound
// from 1.00, so a no change on a noisy row never reads as proof that nothing moved.
func TestNoChangeNamesTheChangeTheRunCouldNotSee(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		low, high float64
		want      string
	}{
		{0.95, 1.05, "none larger than 5.0%"},
		{0.986, 1.023, "none larger than 2.3%"},
		{0.97, 1.01, "none larger than 3.0%"},
	} {
		if got := verdictSentence(noChange, test.low, test.high); !strings.HasSuffix(got, test.want) || !strings.HasPrefix(got, noChange) {
			t.Fatalf("no change over %v to %v read %q; want it to end %q", test.low, test.high, got, test.want)
		}
	}
	if got := verdictSentence("slower", 1.02, 1.05); got != "slower" {
		t.Fatalf("a verdict read %q; want it as it is", got)
	}
}

// Native against Node is a paired ratio with its own interval: a program is a loss only when the whole
// interval sits above 1.00, a win only when it sits below, and within the noise otherwise.
func TestAGapToNodeIsClaimedOnlyOutsideItsInterval(t *testing.T) {
	t.Parallel()
	ratios := []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	withNode := func(factors []float64) programComparison {
		samples := fakeSamples(ratios)
		for index := range samples {
			nodeTime := time.Duration(float64(samples[index][candidateRunner].user) * factors[index])
			samples[index][nodeRunner].user, samples[index][nodeRunner].elapsed = nodeTime, nodeTime
		}
		return compareSamples(samples, "user", fakeFloors, time.Microsecond)
	}
	// Node at half the candidate's time every sample: the candidate is twice Node's, a loss.
	slower := withNode([]float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5})
	if verdictOf(slower.nodeLow, slower.nodeHigh) != "slower" || slower.nodeRatio != 2 {
		t.Fatalf("a candidate at twice Node's time read ratio %v, interval %v to %v; want 2 and slower", slower.nodeRatio, slower.nodeLow, slower.nodeHigh)
	}
	// Node on either side of the candidate: the interval holds 1.00, so neither is claimed.
	mixed := withNode([]float64{0.9, 1.1, 0.95, 1.05, 0.9, 1.1, 0.97, 1.03, 0.9, 1.1, 0.95, 1.05, 1.0, 0.92, 1.08})
	if verdictOf(mixed.nodeLow, mixed.nodeHigh) != noChange {
		t.Fatalf("a candidate on either side of Node read interval %v to %v; want it within the noise", mixed.nodeLow, mixed.nodeHigh)
	}
}

// The plant is the milliseconds it's asked for, whatever the program, and prints nothing.
func TestThePlantIsItsMilliseconds(t *testing.T) {
	t.Parallel()
	loop := plantLoop(95)
	if !strings.Contains(loop, "< 95)") || strings.Contains(loop, "console") {
		t.Fatalf("the plant for 95 ms was %q; want a 95 ms spin that prints nothing", loop)
	}
}
