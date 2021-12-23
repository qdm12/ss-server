package main

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/ss-server/pkg/tcpudp/mock_tcpudp"
)

func Test_Mytest(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // for Go < 1.14

	ctx := context.Background()

	server := mock_tcpudp.NewMockListener(ctrl)
	server.EXPECT().Listen(ctx).Return(nil)

	err := server.Listen(ctx)
	if err != nil {
		t.Error("not expecting an error")
	}
}
