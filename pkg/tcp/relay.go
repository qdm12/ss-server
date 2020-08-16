package tcp

import (
	"io"
	"net"
	"time"
)

// relay copies between left and right connections bidirectionally
func relay(left, right net.Conn, timeNow func() time.Time) (err error) {
	errors := make(chan error)
	defer close(errors)

	copyFn := func(a, b net.Conn, errors chan error) {
		_, copyErr := io.Copy(a, b)
		// wake up the other goroutine blocking on right
		if err := right.SetDeadline(timeNow()); err != nil {
			errors <- err
			return
		}
		// wake up the other goroutine blocking on left
		if err := left.SetDeadline(timeNow()); err != nil {
			errors <- err
			return
		}
		errors <- copyErr
	}

	go copyFn(right, left, errors)
	go copyFn(left, right, errors)

	// Collect eventual errors
	for i := 0; i < 2; i++ {
		copyErr := <-errors
		if copyErr != nil {
			err = copyErr
		}
	}
	return err
}
