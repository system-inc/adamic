package typeaware

import (
	"testing"
	"time"
)

func TestProduct_volume_agreement_oracle(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "oracle")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_archive_checker(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "archive-checker")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_archive_checker_asan(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "archive-checker-asan")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_volume(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "volume")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_volume_asan(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "volume-asan")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_assignable_types(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "assignable-types")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_assignable_types_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "assignable-types-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_widened_shape(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "widened-shape")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_widened_shape_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "widened-shape-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_enum_types(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "enum-types")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_enum_types_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "enum-types-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_scope_locals(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "scope-locals")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_scope_locals_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "scope-locals-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_returns(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-returns")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_returns_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-returns-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_shape(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "property-shape")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_shape_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "property-shape-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_contextual_shape(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "contextual-shape")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_contextual_shape_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "contextual-shape-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_symbol_origin(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "symbol-origin")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_symbol_origin_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "symbol-origin-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_type_origin(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "type-origin")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_type_origin_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "type-origin-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_info(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "property-info")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_info_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "property-info-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_count(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-count")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_count_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-count-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_parameters(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-parameters")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_parameters_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "call-parameters-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_apparent_shape(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "apparent-shape")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_apparent_shape_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "apparent-shape-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_base_shapes(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "base-shapes")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_base_shapes_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "base-shapes-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_handle_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "released-handle-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_registry(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "released-registry")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_registry_native(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "released-registry-native")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}

func TestProduct_volume_agreement_volume_source(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "volume-source")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_source(test *testing.T) {
	test.Parallel()
	started := time.Now()
	volumeAgreementFetch(test, "released-source")
	test.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
