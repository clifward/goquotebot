package storages

import "time"

type AddQuoteRequest struct {
	Author       string
	Content      string
	QuoteContext string
}

type MultipleUnspecifiedQuotesRequest struct {
	QuoteNb int
}
type QuoteResponse struct {
	QuoteID      int
	Author       string
	Content      string
	QuoteContext string
	CreatedAt    time.Time
	DeletedAt    time.Time
	IsActive     bool
	Votes        int
}

type MultipleSpecifiedQuotesRequest struct {
	QuoteIDs []string
}

type UniqueSpecifiedQuoteRequest struct {
	QuoteID int
}

type VoteQuoteRequest struct {
	QuoteID int
	Voter   int64
}

type SearchExpressionRequest struct {
	Expression string
	QuoteNb    int
}
