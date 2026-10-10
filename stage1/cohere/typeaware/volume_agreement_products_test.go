package typeaware

import (
	"testing"
	"time"
)

func TestProduct_volume_agreement_oracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "oracle")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_archive_checker(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "archive-checker")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_archive_checker_asan(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "archive-checker-asan")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_volume(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "volume")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_volume_asan(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "volume-asan")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_assignable_types(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "assignable-types")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_assignable_types_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "assignable-types-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_widened_shape(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "widened-shape")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_widened_shape_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "widened-shape-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_enum_types(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "enum-types")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_enum_types_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "enum-types-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_scope_locals(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "scope-locals")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_scope_locals_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "scope-locals-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_returns(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-returns")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_returns_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-returns-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_shape(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "property-shape")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_shape_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "property-shape-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_contextual_shape(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "contextual-shape")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_contextual_shape_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "contextual-shape-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_symbol_origin(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "symbol-origin")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_symbol_origin_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "symbol-origin-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_type_origin(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "type-origin")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_type_origin_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "type-origin-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_info(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "property-info")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_property_info_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "property-info-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_count(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-count")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_count_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-count-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_parameters(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-parameters")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_call_parameters_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "call-parameters-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_apparent_shape(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "apparent-shape")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_apparent_shape_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "apparent-shape-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_base_shapes(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "base-shapes")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_base_shapes_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "base-shapes-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_handle_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "released-handle-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_registry(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "released-registry")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_registry_native(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "released-registry-native")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}

func TestProduct_volume_agreement_volume_source(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "volume-source")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
func TestProduct_volume_agreement_released_source(t *testing.T) {
	t.Parallel()
	started := time.Now()
	volumeAgreementFetch(t, "released-source")
	t.Logf("product elapsed=%.6fs", time.Since(started).Seconds())
}
