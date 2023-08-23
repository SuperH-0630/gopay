package apple

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SuperH-0630/gopay"
)

// GetAllSubscriptionStatuses Get All Subscription Statuses
// Doc: https://developer.apple.com/documentation/appstoreserverapi/get_all_subscription_statuses
func (c *Client) GetAllSubscriptionStatuses(ctx context.Context, transactionId string) (rsp *AllSubscriptionStatusesRsp, err error) {
	path := fmt.Sprintf(getAllSubscriptionStatuses, transactionId)
	res, bs, err := c.doRequestGet(ctx, path)
	if err != nil {
		return nil, err
	}
	rsp = &AllSubscriptionStatusesRsp{}
	if err = json.Unmarshal(bs, rsp); err != nil {
		return nil, fmt.Errorf("[%w]: %v, bytes: %s", gopay.UnmarshalErr, err, string(bs))
	}
	if res.StatusCode == http.StatusOK {
		return rsp, nil
	}
	if err = statusCodeErrCheck(rsp.StatusCodeErr); err != nil {
		return rsp, err
	}
	return rsp, nil
}
