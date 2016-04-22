package oracle

import "testing"

func TestCommandRegExp(t *testing.T) {
	testStrings := []string{
		"?register",
		"?deadlines",
		"RT @hyfr_ovo: OHHHH HELL NO. @bbykim_  @llrm_  @_Kingmally__  https://t.co/cHJxM9k4yO ?register",
		"RT @5SOS: Are you guys ready for our April fools prank? ?deadlines",
		"foo",
		"",
	}
	expectedResult := []string{
		"?register",
		"?deadlines",
		"?register",
		"?deadlines",
		"",
		"",
	}
	for i, s := range testStrings {
		r := commandRegExp.FindString(s)
		if r != expectedResult[i] {
			t.Errorf("Expected string %s to match %s", s, r)
		}
	}
}

func TestCommandResponse(t *testing.T) {
	testStrings := []string{
		"?register",
		"?deadlines",
		"RT @hyfr_ovo: OHHHH HELL NO. @bbykim_  @llrm_  @_Kingmally__  https://t.co/cHJxM9k4yO ?register",
		"RT @5SOS: Are you guys ready for our April fools prank? ?deadlines",
		"foo",
		"",
	}
	expectedResults := []string{
		registerResponse,
		deadlineResponse,
		registerResponse,
		deadlineResponse,
		unknownRequestResponse,
		unknownRequestResponse,
	}

	for i, s := range testStrings {
		r := responseForMessage(s)
		if r.text != expectedResults[i] {
			t.Errorf("Expected response %s for message %s but saw %s", expectedResults[i], s, r.text)
		}
	}
}

func TestResponseCriteria(t *testing.T) {
	l := 100
	responses := []string{
		deadlineResponse,
		registerResponse,
		helpResponse,
	}

	for _, r := range responses {
		if len(r) > l {
			t.Errorf("Response %s has length greater than 100 chars - actual: %d", r, len(r))
		}
	}
}
