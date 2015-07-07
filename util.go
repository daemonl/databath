package databath

import (
	"regexp"
)

var re_blobjectNotAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9:_ \'-]+`)
var re_notAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var re_numeric = regexp.MustCompile(`^[0-9]*$`)
var re_fieldInSquares = regexp.MustCompile(`\[[a-zA-Z0-9_\.]*\]`)
