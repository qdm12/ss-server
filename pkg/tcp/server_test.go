package tcp

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_closeConnection(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	errTest := errors.New("test error")

	conn := NewMockCloser(ctrl)
	conn.EXPECT().Close().Return(errTest)
	var errs []error

	closeConnection("XYZ", conn, &errs)

	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "closing XYZ: test error")
}
