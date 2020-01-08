package export

import (
	"context"
	"fmt"
)

func (c *Client) SyncSendWithTimeout(ctx context.Context, operator string, param []byte, callback RequestStatusCallback) error {
	ch := make(chan error, 1)
	go func() {
		ch <- c.SyncSend(operator, param, callback)
	}()

	for {
		select {
		case err := <-ch:
			return err
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				err = fmt.Errorf("%s:%w", operator, err)
			}
			return err
		}
	}
}
