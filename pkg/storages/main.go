package storages

type DB interface {
	// Tables
	CreateQuotesTable() error
	DeleteQuotesTable() error
	CreateVotesTable() error
	DeleteVotesTable() error

	// Get
	GetQuotes(MultipleSpecifiedQuotesRequest) ([]QuoteResponse, error)
	GetLastQuotes(MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error)
	GetRandomQuotes(MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error)
	GetTopQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error)
	GetFlopQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error)

	// Add, Delete
	AddQuote(AddQuoteRequest) (string, error)
	DeleteQuote(UniqueSpecifiedQuoteRequest) error

	// Votes
	UpVoteQuote(request VoteQuoteRequest) error
	UnVoteQuote(request VoteQuoteRequest) error
	DownVoteQuote(request VoteQuoteRequest) error

	// Search
	SearchWord(request SearchExpressionRequest) ([]QuoteResponse, error)
	SearchExpression(request SearchExpressionRequest) ([]QuoteResponse, error)

	// DB
	Close() error
}
