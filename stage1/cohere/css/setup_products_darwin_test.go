package css

import "testing"

// The unsanitized parser for leak checking is built on darwin alone (prepareCSSParserCompilerProducts), so its product
// test lives in a darwin file: on linux it would be listed and only skip.
func TestProduct_CSSParserDarwinLeaks(t *testing.T) {
	t.Parallel()
	cssParserPreparedLeaks(t)
}
