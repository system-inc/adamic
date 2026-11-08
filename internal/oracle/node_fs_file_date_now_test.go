package oracle

// The current time: new Date() and Date.now() read CLOCK_REALTIME, and the fixture prints facts about the time, never
// the time (#894zfhq). Mutants: the clock's seconds dropped, or Date.now lowered as 0, and the fixture prints false.
func init() {
	fileSystemFileFixtures = append(fileSystemFileFixtures, "date_now")
}
