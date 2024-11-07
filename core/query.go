package core

import (
	"fmt"
	"time"

	"google.golang.org/api/gmail/v1"
)

func getTodaysEmailsCall(srv *gmail.Service) *gmail.UsersMessagesListCall {
	today := time.Now().Format("2006/01/02")
	return srv.Users.Messages.List("me").Q("after:" + today)
}

func getNonSubscriptionEmailsLastMonthCall(srv *gmail.Service) *gmail.UsersMessagesListCall {
	oneMonthAgo := time.Now().AddDate(0, -1, 0).Format("2006/01/02")
	query := fmt.Sprintf("after:%s -category:{promotions social updates} -label:Automation -label:Subscriptions", oneMonthAgo)
	return srv.Users.Messages.List("me").Q(query)
}
