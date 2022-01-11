package telegram

import (
	"errors"
	c "goquotebot/pkg/storages"
	"testing"
)

func areEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestExtractQuotesID(t *testing.T) {
	samples := []struct {
		Input    string
		Expected []string
	}{
		{
			Input:    "#1",
			Expected: []string{"1"},
		}, {
			Input:    " #1",
			Expected: []string{"1"},
		}, {
			Input:    "#1 ",
			Expected: []string{"1"},
		}, {
			Input:    " #1 ",
			Expected: []string{"1"},
		}, {
			Input:    "#1j",
			Expected: []string{},
		}, {
			Input:    "#1j ",
			Expected: []string{},
		}, {
			Input:    "#1 joij",
			Expected: []string{"1"},
		}, {
			Input:    "hiuh #1 joij",
			Expected: []string{"1"},
		}, {
			Input:    "hiuh #o joij",
			Expected: []string{},
		}, {
			Input:    "  #123 joij",
			Expected: []string{"123"},
		}, {
			Input:    "  #1000joij",
			Expected: []string{},
		}, {
			Input:    "200",
			Expected: []string{},
		}, {
			Input:    "  #pok #123 #987 zhhui #98a #JJ #908 joi#12 #9",
			Expected: []string{"123", "987", "908", "9"},
		}, {
			Input:    "#987 #123",
			Expected: []string{"987", "123"},
		},
	}

	for _, sample := range samples {
		tmp := ExtractQuotesID(sample.Input)
		if !areEquals(tmp, sample.Expected) {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestExtractQuote(t *testing.T) {
	samples := []struct {
		Input    string
		Expected []string
	}{
		{
			Input:    "/add  test | hello",
			Expected: []string{"test", "hello"},
		}, {
			Input:    "/add |test|",
			Expected: []string{},
		}, {
			Input:    "/add|test| cc|",
			Expected: []string{},
		}, {
			Input:    "/add    test|cc",
			Expected: []string{"test", "cc"},
		}, {
			Input:    "/add A long quote|owner",
			Expected: []string{"A long quote", "owner"},
		}, {
			Input:    "/add A quote without owner",
			Expected: []string{},
		}, {
			Input:    "/add A quote without owner |",
			Expected: []string{},
		}, {
			Input:    "/add I want to love Go a lot but your algorithm is blocking me | Gopher",
			Expected: []string{"I want to love Go a lot but your algorithm is blocking me", "Gopher"},
		},
	}

	for _, sample := range samples {
		tmp := ExtractQuote(sample.Input)
		if !areEquals(tmp, sample.Expected) {
			t.Errorf("got %q, wanted %q for %s", tmp, sample.Expected, sample.Input)
		}
	}
}

func TestExtractNumber(t *testing.T) {
	samples := []struct {
		Input         string
		ErrorExpected error
		Expected      int
	}{
		{
			Input:         "/last",
			ErrorExpected: nil,
			Expected:      1,
		}, {
			Input:         "/random 30",
			ErrorExpected: nil,
			Expected:      30,
		}, {
			Input:         "/last #30  ",
			ErrorExpected: nil,
			Expected:      30,
		}, {
			Input:         "/last #Q30  ",
			ErrorExpected: nil,
			Expected:      30,
		}, {
			Input:         "/hello Q30  ",
			ErrorExpected: nil,
			Expected:      30,
		}, {
			Input:         "/last joij",
			ErrorExpected: nil,
			Expected:      1,
		},
	}

	for _, sample := range samples {
		tmp, err := ExtractNumber(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestExtractID(t *testing.T) {
	samples := []struct {
		Input         string
		ErrorExpected error
		Expected      int
	}{
		{
			Input:         "/command 10",
			ErrorExpected: nil,
			Expected:      10,
		}, {
			Input:         "/delete Q30",
			ErrorExpected: nil,
			Expected:      30,
		}, {
			Input:         "/upvote #Q20  ",
			ErrorExpected: nil,
			Expected:      20,
		}, {
			Input:         "/downvote #40  ",
			ErrorExpected: nil,
			Expected:      40,
		}, {
			Input:         "/hello  ",
			ErrorExpected: ErrNoIDProvided,
		}, {
			Input:         "/last joij",
			ErrorExpected: ErrNoIDProvided,
		},
	}

	for _, sample := range samples {
		tmp, err := ExtractID(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v for the input : %s", err, sample.ErrorExpected, sample.Input)
			continue
		}
		if sample.ErrorExpected == nil && tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestExtractSearchExpressionRequest(t *testing.T) {
	samples := []struct {
		Input         string
		ErrorExpected error
		Expected      c.SearchExpressionRequest
	}{
		{
			Input:         "/search 10",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{Expression: "10", QuoteNb: 1},
		}, {
			Input:         "/search  a short|expression  10 ",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{Expression: "a short|expression", QuoteNb: 10},
		}, {
			Input:         "/search",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{},
		}, {
			Input:         "/search word",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{Expression: "word", QuoteNb: 1},
		}, {
			Input:         "/search expression   2",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{Expression: "expression", QuoteNb: 2},
		}, {
			Input:         "/searchword hello ",
			ErrorExpected: nil,
			Expected:      c.SearchExpressionRequest{Expression: "hello", QuoteNb: 1},
		},
	}

	for _, sample := range samples {
		tmp := ExtractExpressionAndNumber(sample.Input)
		if tmp.Expression != sample.Expected.Expression || tmp.QuoteNb != sample.Expected.QuoteNb {
			t.Errorf("got %v, wanted %v for input %s", tmp, sample.Expected, sample.Input)
		}
	}
}

func TestGenerateQuotesMessage(t *testing.T) {
	samples := []struct {
		Input         []c.QuoteResponse
		ErrorExpected error
		Expected      string
	}{
		{
			Input: []c.QuoteResponse{
				{
					QuoteID:      4,
					Author:       "Author",
					Content:      "Content of the quote",
					QuoteContext: "Contexte",
					Votes:        2,
				},
			},
			ErrorExpected: nil,
			Expected:      "\n#Q4 (+2)\n*Content of the quote*\n\n_by Contexte_",
		},
		{
			Input: []c.QuoteResponse{
				{
					QuoteID:      4,
					Author:       "Author 1",
					Content:      "Content 1",
					QuoteContext: "Contexte 1",
					Votes:        3,
				},
				{
					QuoteID:      5,
					Author:       "Author 2",
					Content:      "Content 2 | \n Test",
					QuoteContext: "Contexte 2<>",
					Votes:        -4,
				},
			},
			ErrorExpected: nil,
			Expected:      "\n#Q4 (+3)\n*Content 1*\n\n_by Contexte 1_\n\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\\_\n\n#Q5 (-4)\n*Content 2 | \n Test*\n\n_by Contexte 2<>_",
		},
	}

	for _, sample := range samples {
		tmp, err := GenerateQuotesMessage(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestGenerateNewQuoteMessage(t *testing.T) {
	samples := []struct {
		Input         c.AddQuoteRequest
		ErrorExpected error
		Expected      string
	}{
		{
			Input: c.AddQuoteRequest{
				Author:       "9080987",
				Content:      "<oij!jmoij>",
				QuoteContext: "jj!|&$ù",
			},
			ErrorExpected: nil,
			Expected:      "✅ New quote added ✅\n*<oij!jmoij>*\n\n_by jj!|&$ù_\n",
		},
	}

	for _, sample := range samples {
		tmp, err := GenerateNewQuoteMessage(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestGenerateDeleteQuoteMessage(t *testing.T) {
	samples := []struct {
		Input         c.UniqueSpecifiedQuoteRequest
		ErrorExpected error
		Expected      string
	}{
		{
			Input:         c.UniqueSpecifiedQuoteRequest{QuoteID: 4},
			ErrorExpected: nil,
			Expected:      "✅ Quote deleted : #Q4\n",
		},
	}

	for _, sample := range samples {
		tmp, err := GenerateDeleteQuoteMessage(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestGenerateVoteAddedMessage(t *testing.T) {
	samples := []struct {
		Input         c.VoteQuoteRequest
		ErrorExpected error
		Expected      string
	}{
		{
			Input:         c.VoteQuoteRequest{QuoteID: 12},
			ErrorExpected: nil,
			Expected:      "✅ *Vote Registered ✅\nYour vote about the quote #Q12 has been registered.",
		},
	}

	for _, sample := range samples {
		tmp, err := GenerateVoteAddedMessage(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestGenerateVoteRemovedMessage(t *testing.T) {
	samples := []struct {
		Input         c.VoteQuoteRequest
		ErrorExpected error
		Expected      string
	}{
		{
			Input:         c.VoteQuoteRequest{QuoteID: 15},
			ErrorExpected: nil,
			Expected:      "✅ *Vote removed* ✅\nYou have successfully removed your vote on the quote #Q15.",
		},
	}

	for _, sample := range samples {
		tmp, err := GenerateVoteRemovedMessage(sample.Input)
		if err != sample.ErrorExpected {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}

func TestConvertMatchToInt(t *testing.T) {
	samples := []struct {
		Input         []string
		ErrorExpected error
		Expected      int
	}{
		{
			Input:         []string{"", "1"},
			ErrorExpected: nil,
			Expected:      1,
		}, {
			Input:         []string{"", "10"},
			ErrorExpected: nil,
			Expected:      10,
		}, {
			Input:         []string{"", "10e"},
			ErrorExpected: errors.New("strconv.Atoi: parsing \"10e\": invalid syntax"),
			Expected:      0,
		},
	}

	for _, sample := range samples {
		tmp, err := ConvertMatchToInt(sample.Input)
		if err != nil && err.Error() != sample.ErrorExpected.Error() {
			t.Errorf("got %v instead of %v", err, sample.ErrorExpected)
			continue
		}
		if tmp != sample.Expected {
			t.Errorf("got %q, wanted %q", tmp, sample.Expected)
		}
	}
}
