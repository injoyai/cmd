package resource

import "os"

func lockCache(filename string) (func(), error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		os.Remove(filename)
		return nil, err
	}
	return func() { os.Remove(filename) }, nil
}
