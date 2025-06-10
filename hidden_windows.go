//go:build windows

package den

import "syscall"

func isHiddenFile(name string) (bool, error) {
	pointer, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
