package esl

const (
	apiResponse      = "api/response"
	authRequest      = "auth/request"
	commandReply     = "command/reply"
	logData          = "log/data"
	disconnectNotice = "text/disconnect-notice"
	eventPlain       = "text/event-plain"
	// rejected by acl
	rudeRejection = "text/rude-rejection"

	replyOK  = "+OK"
	replyERR = "-ERR"
)
