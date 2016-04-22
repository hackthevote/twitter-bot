package bot

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"log"
	"sync"
	"voteinfobot/oracle"
)

const twitterBotID = 715540629829369857

type twitterResponse struct {
	msg          *oracle.Response
	requestTweet *twitter.Tweet
}

func (tr *twitterResponse) String() string {
	return fmt.Sprintf("%s response message for tweet %d", tr.msg, tr.requestTweet.ID)
}

type twitterHandler struct {
	mu        sync.Mutex
	masterWg  *sync.WaitGroup
	twitterWg *sync.WaitGroup
	client    *twitter.Client
	stream    *twitter.Stream
	responseC chan *twitterResponse
	logC      chan string
}

func (th *twitterHandler) String() string {
	return "This is a twitterHandler"
}

func (th *twitterHandler) startActivity() {
	th.twitterWg.Add(1)
}

func (th *twitterHandler) endActivity() {
	th.twitterWg.Done()
}

func (th *twitterHandler) logMsg(m string) {
	th.startActivity()
	go func() {
		th.logC <- m
		th.endActivity()
	}()
}

func (th *twitterHandler) sendTweet(r *twitterResponse) {
	tweetContent := genTweetTextForMsg(r)
	statusParams := new(twitter.StatusUpdateParams)

	// The request tweet is in reply to another user - we should send our reply to
	// that user's tweet and mention the requester.  If the req tweet is in reply to the bot's own tweets
	// we should reply to the requester directly.
	if r.requestTweet.User.ID != r.requestTweet.InReplyToUserID && r.requestTweet.InReplyToUserID != twitterBotID {
		statusParams.InReplyToStatusID = r.requestTweet.InReplyToStatusID
	} else {
		// otherwise send our reply to requester
		statusParams.InReplyToStatusID = r.requestTweet.ID
	}

	t, _, err := th.client.Statuses.Update(tweetContent, statusParams)
	if err != nil {
		th.logMsg(err.Error())
	}
	th.logMsg(fmt.Sprintf("[TWITTER_RESPONSE] [CONTENT: %s] [TWEET_ID: %d] [REPLIED_TO_TWEET_ID: %d] [REPLIED_TO_USER_ID: %d]", t.Text, t.ID, t.InReplyToStatusID, t.InReplyToUserID))
	th.endActivity()
}

func (th *twitterHandler) handleReplies() {
	for r := range th.responseC {
		th.startActivity()
		go th.sendTweet(r)
		th.endActivity()
	}
}

func (th *twitterHandler) Start() {
	th.logMsg("[TWITTER] bot starting")
	th.mu.Lock()
	defer th.mu.Unlock()

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(t *twitter.Tweet) {
		// We don't want to reply to ourselves or respond to a retweet
		if t.User.ID != twitterBotID && t.RetweetedStatus == nil {
			// TODO presently no need for this goroutine/channel until such a point as the response request
			// might block - could simplify down to a call to sendTweet
			msg := oracle.Consult(t.Text)
			// On Twitter the bot could be mentioned many times without actually being asked anything
			// and that would lead to the bot spamming its help message
			if msg.Status == oracle.Answered {
				th.logMsg(fmt.Sprintf("[TWITTER_REQUEST] [CONTENT: %s] [TWEET_ID: %d] [TWITTER_USER: %d]", t.Text, t.ID, t.User.ID))
				th.startActivity()
				go func() {
					th.responseC <- &twitterResponse{
						msg:          msg,
						requestTweet: t,
					}
				}()
			} else {
				th.logMsg(fmt.Sprintf("[TWITTER_NO_ANSWER] [CONTENT: %s] [TWEET_ID: %d] [TWITTER_USER: %d", t.Text, t.ID, t.User.ID))
			}
		}
	}

	params := &twitter.StreamUserParams{
		With:          "user",
		StallWarnings: twitter.Bool(true),
	}

	stream, err := th.client.Streams.User(params)
	if err != nil {
		log.Panic(err.Error())
	}
	th.stream = stream

	go demux.HandleChan(th.stream.Messages)
	go th.handleReplies()
}

func (th *twitterHandler) Stop() {
	th.logMsg("[TWITTER] bot stopping")
	th.mu.Lock()
	defer th.mu.Unlock()
	th.stream.Stop()
	th.twitterWg.Wait()
	close(th.responseC)
	th.masterWg.Done()
}

func genTweetTextForMsg(r *twitterResponse) string {
	var userMentions string

	// We've been asked to reply to another user but we don't want to reply to a bot tweet.
	if r.requestTweet.InReplyToUserID != r.requestTweet.User.ID && r.requestTweet.InReplyToUserID != twitterBotID {
		userMentions = fmt.Sprintf("@%s @%s", r.requestTweet.InReplyToScreenName, r.requestTweet.User.ScreenName)
	} else {
		userMentions = "@" + r.requestTweet.User.ScreenName
	}

	return userMentions + " " + r.msg.Text
}

// NewTwitterHandler returns a properly configured
// twitterHandler
func NewTwitterHandler(wg *sync.WaitGroup, logC chan string, cKey string, cSecKey string, aTok string, aSecTok string) *twitterHandler {
	config := oauth1.NewConfig(cKey, cSecKey)
	token := oauth1.NewToken(aTok, aSecTok)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	return &twitterHandler{
		masterWg:  wg,
		client:    client,
		responseC: make(chan *twitterResponse),
		twitterWg: &sync.WaitGroup{},
		logC:      logC,
	}
}
