package javascript

import "testing"

func TestFX7ViewCycle(t *testing.T) {
	t.Parallel()
	source := viewTestRuntime + viewUntaggedPlainRuntime + `
 const adamicFieldReadiness=new WeakMap();
 const value={};value.next=value;
 const contracts=[{Kind:2,Fields:[{Name:'next',Contract:2}]},{Kind:4,Members:[1,3]},{Kind:7}];
 console.log(adamicUntaggedPlainSelect(value,contracts,2,'view.value','Link | undefined'));
 `
	runViewNode(t, source, "1\n", "", 0)
}

func TestFX7TupleMarkerRejectsOrdinaryArray(t *testing.T) {
	t.Parallel()
	source := viewTestRuntime + fieldReadinessRuntime + viewUntaggedPlainRuntime + `
 const contracts=[{Kind:4,Members:[2]},{Kind:2,FixedTuple:true,Tuple:[3,3],Fields:[{Name:'0',Contract:3},{Name:'1',Contract:3}]},{Kind:1,Of:1}];
 console.log(adamicUntaggedPlainSelect([1,2],contracts,1,'view.value','Pair'));
 `
	runViewNode(t, source, "", "adamic: panic: field read failed: view.value matches no member of Pair; expected Pair, found array\n", 70)
}
