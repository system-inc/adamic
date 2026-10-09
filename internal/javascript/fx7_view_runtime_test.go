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
