package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/ss-server/pkg/tcpudp/mock_tcpudp"
)

func Test(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // for Go < 1.14
	server := mock_tcpudp.NewMockListener(ctrl)
	server.EXPECT().Listen(context.Background(), ":8388")
	// more of your test using server
}
