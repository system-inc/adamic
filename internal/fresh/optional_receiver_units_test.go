package fresh_test

import "testing"

func init() {
	freshCorpusPaths = append(freshCorpusPaths, "../oracle/testdata/library_optional_receiver_delete.a")
	freshCorpusTests[0] = append(freshCorpusTests[0], TestFreshWrites____oracle_testdata_library_optional_receiver_delete_a_f3ecf551fb04)
	freshCorpusPaths = append(freshCorpusPaths, "../oracle/testdata/library_optional_receiver_foreach.a")
	freshCorpusTests[0] = append(freshCorpusTests[0], TestFreshWrites____oracle_testdata_library_optional_receiver_foreach_a_dacec7d03e89)
	freshCorpusPaths = append(freshCorpusPaths, "../oracle/testdata/library_optional_receiver_effects.a")
	freshCorpusTests[0] = append(freshCorpusTests[0], TestFreshWrites____oracle_testdata_library_optional_receiver_effects_a_d4b1b369096a)
	freshCorpusPaths = append(freshCorpusPaths, "../oracle/testdata/library_optional_receiver_result.a")
	freshCorpusTests[0] = append(freshCorpusTests[0], TestFreshWrites____oracle_testdata_library_optional_receiver_result_a_7b753efbbcc5)
}

func TestFreshWrites____oracle_testdata_library_optional_receiver_delete_a_f3ecf551fb04(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/library_optional_receiver_delete.a")
}

func TestFreshWrites____oracle_testdata_library_optional_receiver_foreach_a_dacec7d03e89(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/library_optional_receiver_foreach.a")
}

func TestFreshWrites____oracle_testdata_library_optional_receiver_effects_a_d4b1b369096a(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/library_optional_receiver_effects.a")
}

func TestFreshWrites____oracle_testdata_library_optional_receiver_result_a_7b753efbbcc5(t *testing.T) {
	t.Parallel()
	beginFreshUnit(t)
	checkFreshProgram(t, "../oracle/testdata/library_optional_receiver_result.a")
}
