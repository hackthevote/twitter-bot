package oracle

import (
	"fmt"
	"regexp"
)

// TODO load responses from file, possible cmd line switch
const (
	commands               = `\?(deadlines|register|help)`
	deadlineResponse       = "Upcoming UK deadlines - register by: 7th June for EU referendum"
	registerResponse       = "Registering to vote in the UK is easy and fast - just go to https://www.gov.uk/register-to-vote"
	helpResponse           = "Avail commands: ?deadlines | ?register | ?help"
	unknownRequestResponse = `Sorry - VoteInfoBot didn't understand your query - try ?help for commands`
)

type responseStatus int

const (
	UnansweredNoCmdFound responseStatus = iota
	Answered
)

var commandRegExp *regexp.Regexp
var commandResponses = map[string]string{
	"?deadlines": deadlineResponse,
	"?register":  registerResponse,
	"?help":      helpResponse,
}


type Response struct {
	Status responseStatus
	Text   string
	Cmd    string
}

func (or *Response) String() string {
	var status string
	switch or.Status {
	case UnansweredNoCmdFound:
		status = "No command found in message"
	case Answered:
		status = fmt.Sprintf("Command %s", or.Cmd)
	}
	return fmt.Sprintf("%s with response %s", status, or.Text)
}

func getCommandForMsg(msg string) string {
	return commandRegExp.FindString(msg)
}

// Consult returns an appropriate response for the provided
// text.  Consult the oracle.
func Consult(text string) *Response {
	c := getCommandForMsg(text)
	r := commandResponses[c]
	var s responseStatus
	if r == "" {
		s = UnansweredNoCmdFound
		r = unknownRequestResponse
	} else {
		s = Answered
	}
	return &Response{
		Status: s,
		Text:   r,
		Cmd:    c,
	}
}

func init() {
	commandRegExp = regexp.MustCompile(commands)
}
