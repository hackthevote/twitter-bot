package bot

import (
	"testing"
	"github.com/dghubble/go-twitter/twitter"
)

func TestGenTweetForMsg(t *testing.T) {

	testReqTweet := &twitterResponse{
		msg: &oracleResponse{
			status: answered,
			text: deadlineResponse,
			cmd: "?deadline",
		},
		requestTweet: &twitter.Tweet{
			User: &twitter.User{
				ID: 1,
				ScreenName: "requester",
			},
			InReplyToScreenName: "replied_to",
			InReplyToUserID: 2,
		},
	}

	// if request is itself a reply to another user, make sure they are mentioned
	expectTweetText := "@requester @replied_to " + deadlineResponse

	tweetText := genTweetTextForMsg(testReqTweet)
	if tweetText != expectTweetText {
		t.Errorf("expected genned tweet text to be %s but saw %s", expectTweetText, tweetText)
	}

	testReqTweet.requestTweet.InReplyToScreenName = "requester"
	testReqTweet.requestTweet.InReplyToUserID = 1

	// if request is a reply to the requester, they should be mentioned only once
	expectTweetText = "@requester " + deadlineResponse
	tweetText = genTweetTextForMsg(testReqTweet)
	if tweetText != expectTweetText {
		t.Errorf("expected genned tweet text to be %s but saw %s", expectTweetText, tweetText)
	}

	// if request is a reply to the twitter bot then the twitter bot should not mention itself
	testReqTweet.requestTweet.InReplyToScreenName = "voteinfobot"
	testReqTweet.requestTweet.InReplyToUserID = twitterBotID

	expectTweetText = "@requester " + deadlineResponse
	tweetText = genTweetTextForMsg(testReqTweet)
	if tweetText != expectTweetText {
		t.Errorf("expected genned tweet text to be %s but saw %s", expectTweetText, tweetText)
	}
}
