package ioutil

import "io"

type MultiCloser []io.Closer

func (s MultiCloser) Close() (err error) {
	for _, cl := range s {
		if err1 := cl.Close(); err == nil && err1 != nil {
			err = err1
		}
	}
	return
}
