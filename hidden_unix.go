//go:build !windows

package den

func isHiddenFile(name string) (bool, error) {
	return len(name) > 1 && name[0] == '.', nil
}
