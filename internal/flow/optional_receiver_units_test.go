package flow

import "testing"

func init() {
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/library_optional_receiver_delete.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_library_optional_receiver_delete_a_f3ecf551fb04)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/library_optional_receiver_foreach.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_library_optional_receiver_foreach_a_dacec7d03e89)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/library_optional_receiver_effects.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_library_optional_receiver_effects_a_d4b1b369096a)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/library_optional_receiver_result.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_library_optional_receiver_result_a_7b753efbbcc5)
	}
}

func TestFlowProgram____oracle_testdata_library_optional_receiver_delete_a_f3ecf551fb04(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/library_optional_receiver_delete.a")
}

func TestFlowProgram____oracle_testdata_library_optional_receiver_foreach_a_dacec7d03e89(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/library_optional_receiver_foreach.a")
}

func TestFlowProgram____oracle_testdata_library_optional_receiver_effects_a_d4b1b369096a(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/library_optional_receiver_effects.a")
}

func TestFlowProgram____oracle_testdata_library_optional_receiver_result_a_7b753efbbcc5(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/library_optional_receiver_result.a")
}
