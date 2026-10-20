// A package named filepath at another path: its WalkDir and Glob are not the ones the rule knows.
package filepath

import "io/fs"

func WalkDir(root string, visit fs.WalkDirFunc) error { return nil }

func Glob(pattern string) ([]string, error) { return nil, nil }
