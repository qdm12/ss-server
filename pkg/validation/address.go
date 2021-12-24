package validation

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

var (
	ErrListenAddressNotValid = errors.New("listening address is not valid")
	ErrListenPortNotValid    = errors.New("listening port is not valid")
	ErrListenPortPrivileged  = errors.New("cannot use a privileged listening port without running as root")
)

func ValidateAddress(address string) (err error) {
	_, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListenAddressNotValid, err)
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListenPortNotValid, err)
	}
	if portInt < 0 || portInt > 65535 {
		return fmt.Errorf("%w: %d: must be between 0 and 65535",
			ErrListenPortNotValid, portInt)
	}

	uid := os.Getuid()
	const maxPrivilegedPort = 1023
	if portInt <= maxPrivilegedPort && uid != 0 {
		return fmt.Errorf("%w: port %d with user id %d",
			ErrListenPortPrivileged, portInt, uid)
	}

	return nil
}
