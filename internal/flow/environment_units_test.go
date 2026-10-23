package flow

import (
	"path/filepath"
	"testing"
)

// Own each environment fixture as a bounded unit using the current corpus
// declarations and checks. Historical generated units keep their indices.
func init() {
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_array.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_array_a_0e055bb41e78)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_async.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_async_a_8176bcb60b75)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_callback_escape.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_callback_escape_a_07d564404a0d)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_callbacks.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_callbacks_a_d8cee65434a8)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_capture.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_capture_a_5421acbf403a)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_captured_parameter_escape.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_captured_parameter_escape_a_569aafab78bc)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_captured_parameter_reassign.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_captured_parameter_reassign_a_666598e2453b)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_captured_parameters.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_captured_parameters_a_b780a7b48fd0)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_alias_join.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_alias_join_a_2f8a067d4dd0)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_boundary64.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_boundary64_a_b57831c8375e)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_bounded_call.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_bounded_call_a_2a87b762c692)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_callback_exits.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_callback_exits_a_1e64cf73c6f4)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_callback_result.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_callback_result_a_e49466ed7914)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_callbacks_extra.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_callbacks_extra_a_4b79f6e4ed4f)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_captured_parameter.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_captured_parameter_a_49e5705d94c7)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_closure_self.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_closure_self_a_c0d80c8b36ca)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_collections.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_collections_a_4bf77ec89e13)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_initializer_throw.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_initializer_throw_a_bde111a8d20b)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_local_containers.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_local_containers_a_ddec1ff9b8e9)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_primitives.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_primitives_a_69a8e83ae097)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_region_values.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_region_values_a_4b2e8b2e9142)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_returned_values.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_returned_values_a_5bf56442570d)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_coverage_values_exits.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_coverage_values_exits_a_508c02d732be)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_direct.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_direct_a_8ed366904b7f)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_exits.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_exits_a_5a0ca5eb2be2)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_field.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_field_a_7a4256284e41)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_global.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_global_a_0d6039193c6e)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_keeping_call.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_keeping_call_a_cd762895c22a)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_large.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_large_a_ba86c267160c)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_local_call.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_local_call_a_7b240654397f)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_loop.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_loop_a_a4dc8d957131)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_map.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_map_a_1ab1662b5b18)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_returned.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_returned_a_7754e385a1f6)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_set.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_set_a_0a2fce7234d6)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_siblings.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_siblings_a_50e41100269f)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_unknown_call.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_unknown_call_a_546fa412192e)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_unknown_callback.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_unknown_callback_a_d94a97336fb8)
	}
	flowCorpusPaths = append(flowCorpusPaths, "../oracle/testdata/environment_virtual_call.a")
	for family := range flowCorpusTests {
		flowCorpusTests[family] = append(flowCorpusTests[family], TestFlowProgram____oracle_testdata_environment_virtual_call_a_4ee3b36e7269)
	}
}

func TestFlowProgram____oracle_testdata_environment_array_a_0e055bb41e78(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_array.a")
}
func TestFlowProgram____oracle_testdata_environment_async_a_8176bcb60b75(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_async.a")
}
func TestFlowProgram____oracle_testdata_environment_callback_escape_a_07d564404a0d(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_callback_escape.a")
}
func TestFlowProgram____oracle_testdata_environment_callbacks_a_d8cee65434a8(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_callbacks.a")
}
func TestFlowProgram____oracle_testdata_environment_capture_a_5421acbf403a(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_capture.a")
}
func TestFlowProgram____oracle_testdata_environment_captured_parameter_escape_a_569aafab78bc(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_captured_parameter_escape.a")
}
func TestFlowProgram____oracle_testdata_environment_captured_parameter_reassign_a_666598e2453b(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_captured_parameter_reassign.a")
}
func TestFlowProgram____oracle_testdata_environment_captured_parameters_a_b780a7b48fd0(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_captured_parameters.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_alias_join_a_2f8a067d4dd0(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_alias_join.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_boundary64_a_b57831c8375e(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_boundary64.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_bounded_call_a_2a87b762c692(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_bounded_call.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_callback_exits_a_1e64cf73c6f4(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_callback_exits.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_callback_result_a_e49466ed7914(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_callback_result.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_callbacks_extra_a_4b79f6e4ed4f(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_callbacks_extra.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_captured_parameter_a_49e5705d94c7(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_captured_parameter.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_closure_self_a_c0d80c8b36ca(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_closure_self.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_collections_a_4bf77ec89e13(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_collections.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_initializer_throw_a_bde111a8d20b(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_initializer_throw.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_local_containers_a_ddec1ff9b8e9(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_local_containers.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_primitives_a_69a8e83ae097(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_primitives.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_region_values_a_4b2e8b2e9142(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_region_values.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_returned_values_a_5bf56442570d(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_returned_values.a")
}
func TestFlowProgram____oracle_testdata_environment_coverage_values_exits_a_508c02d732be(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_coverage_values_exits.a")
}
func TestFlowProgram____oracle_testdata_environment_direct_a_8ed366904b7f(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_direct.a")
}
func TestFlowProgram____oracle_testdata_environment_exits_a_5a0ca5eb2be2(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_exits.a")
}
func TestFlowProgram____oracle_testdata_environment_field_a_7a4256284e41(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_field.a")
}
func TestFlowProgram____oracle_testdata_environment_global_a_0d6039193c6e(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_global.a")
}
func TestFlowProgram____oracle_testdata_environment_keeping_call_a_cd762895c22a(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_keeping_call.a")
}
func TestFlowProgram____oracle_testdata_environment_large_a_ba86c267160c(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_large.a")
}
func TestFlowProgram____oracle_testdata_environment_local_call_a_7b240654397f(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_local_call.a")
}
func TestFlowProgram____oracle_testdata_environment_loop_a_a4dc8d957131(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_loop.a")
}
func TestFlowProgram____oracle_testdata_environment_map_a_1ab1662b5b18(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_map.a")
}
func TestFlowProgram____oracle_testdata_environment_returned_a_7754e385a1f6(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_returned.a")
}
func TestFlowProgram____oracle_testdata_environment_set_a_0a2fce7234d6(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_set.a")
}
func TestFlowProgram____oracle_testdata_environment_siblings_a_50e41100269f(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_siblings.a")
}
func TestFlowProgram____oracle_testdata_environment_unknown_call_a_546fa412192e(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_unknown_call.a")
}
func TestFlowProgram____oracle_testdata_environment_unknown_callback_a_d94a97336fb8(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_unknown_callback.a")
}
func TestFlowProgram____oracle_testdata_environment_virtual_call_a_4ee3b36e7269(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkAllFlowProgram(t, "../oracle/testdata/environment_virtual_call.a")
}

func TestEnvironmentFlowUnitsCoverFixtures(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("../oracle/testdata/environment_*.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("environment fixture inventory is empty")
	}
	for _, path := range paths {
		found := 0
		for _, registered := range flowCorpusPaths {
			if registered == filepath.ToSlash(path) {
				found++
			}
		}
		if found != 1 {
			t.Fatalf("environment fixture needs exactly one bounded unit: %s has %d", path, found)
		}
	}
}
