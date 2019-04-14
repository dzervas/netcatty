package action

import "github.com/amoghe/distillog"

var Log = distillog.NewStdoutLogger("action")

var State = map[string]string{}
