// Copyright 2014 Simon Zimmermann. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package util/sig is simple sig trap closer
*/

package sig

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/simonz05/util/log"
)

type Cleanup func() error

func sigTrapCloser(l net.Listener, cleanups ...Cleanup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for _ = range c {
			for _, cb := range cleanups {
				if err := cb(); err != nil {
					log.Error(err)
				}
			}
			l.Close()
			log.Printf("Closed listener %s", l.Addr())
		}
	}()
}
