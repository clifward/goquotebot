package storages

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestNewSqliteWrapper(t *testing.T) {
	fmt.Println("TestNewSqliteWrapper")
	server, err := NewSqliteWrapper("/tmp/test.db")
	if err != nil {
		t.Errorf("error %v should not have occured", err)
	}
	err = server.Close()
	if err != nil {
		t.Errorf("error %v should not have occured", err)
	}
}

func TestMeasureIndex(t *testing.T) {
	threshold := 0.6
	samples := []struct {
		Input            string
		ExistingContents map[int]string
		ExpectedBool     bool
	}{
		{
			Input: "A very simple text",
			ExistingContents: map[int]string{
				1: "A very simple text",
				2: "with another message",
			},
			ExpectedBool: true,
		}, {
			Input: "A very simple quote",
			ExistingContents: map[int]string{
				1: "A simple quote",
				2: "there is a missing word"},
			ExpectedBool: true,
		}, {
			Input: "A quote among others",
			ExistingContents: map[int]string{
				1: "I love to quote people",
				2: "The sun is a star among others",
				3: "I love art !"},
			ExpectedBool: false,
		},
		{
			Input: "I miss my book",
			ExistingContents: map[int]string{
				1: "I miss my colleagues",
				2: "They are so fun"},
			ExpectedBool: false,
		},
		{
			Input: "I miss my computer",
			ExistingContents: map[int]string{
				1: "I miss my contuter"},
			ExpectedBool: true,
		},
		{
			Input: "I miss my computer, I forgot it at my house and I'm sad now",
			ExistingContents: map[int]string{
				1: "I forgot it at my house and I'm sad now"},
			ExpectedBool: true,
		},
		{
			Input: "hello",
			ExistingContents: map[int]string{
				1: "I said hello to my brother, but he didn't answer",
				2: "I named my dog \"Hello\"",
				3: "Hello My Old Friend !"},
			ExpectedBool: false,
		},
		{
			Input: "Je n'étais pas là",
			ExistingContents: map[int]string{
				1: "J'etais pas la"},
			ExpectedBool: true,
		},
		{
			Input: "A: 'raconte un truc marrant'\nB: 'clash A de façon énervée'",
			ExistingContents: map[int]string{
				1: "'clash A de façon énervée'"},
			ExpectedBool: true,
		},
		{
			Input: "Une quote en français avec des caractères un peu spéciaux: il en faut",
			ExistingContents: map[int]string{
				1: "d'autres quotes qui n'ont rien à voir",
				2: "une autre, mais pareil, aucun rapport"},
			ExpectedBool: false,
		},
		{
			Input: "Une quote en français avec des caractères un peu spéciaux: il en faut",
			ExistingContents: map[int]string{
				1: "Une citation en français avec des caractères un peu spéciaux: y en faut",
				2: "une autre, mais pareil, aucun rapport"},
			ExpectedBool: true,
		},
	}

	for _, sample := range samples {
		isProbablyStored, _, _ := measureIndex(sample.Input, sample.ExistingContents, threshold)
		if isProbablyStored != sample.ExpectedBool {
			t.Errorf("error during measureIndex for %s, expecting %t", sample.Input, sample.ExpectedBool)
		}
	}
}

func TestCheckIfExists(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := AddQuoteRequest{
		Author:       "author",
		Content:      "content",
		QuoteContext: "context",
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < 5; i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}
	query := "SELECT .*? FROM Quotes WHERE Quotes.isAvailable=true ORDER BY Quotes.quoteID DESC LIMIT 5 "
	mock.ExpectQuery(query).WillReturnRows(rows)
	isStored, similarQuotes, _, err := w.checkIfExists(request)
	if err != nil {
		t.Errorf("Error in CheckIfExists: %v", err)
	}
	_, _ = isStored, similarQuotes
}

func TestAddQuote(t *testing.T) {

	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}

	for i := 0; i < 20; i++ {
		quote := AddQuoteRequest{
			Author:       fmt.Sprintf("author%d", rand.Intn(10)+1),
			Content:      fmt.Sprintf("blabla content n°%d", rand.Intn(10)+1),
			QuoteContext: fmt.Sprintf("context%d", rand.Intn(10)+1),
		}

		rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
		for i := 0; i < 5; i++ {
			u := QuoteResponse{
				QuoteID:      i,
				Author:       "a",
				Content:      "b",
				QuoteContext: "c",
				CreatedAt:    time.Time{},
				DeletedAt:    time.Time{},
				IsActive:     true,
				Votes:        1,
			}
			rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
		}
		query := "SELECT .*? FROM Quotes WHERE Quotes.isAvailable=true ORDER BY Quotes.quoteID DESC LIMIT 5 "
		mock.ExpectQuery(query).WillReturnRows(rows)

		query = "INSERT INTO Quotes \\(content, context, author, createdAt, isAvailable\\) VALUES \\(.*?,.*?,.*?,CURRENT_TIMESTAMP,.*?\\)"
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.Content, quote.QuoteContext, quote.Author, 1).WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := w.AddQuote(quote)
		if err != nil {
			t.Errorf("Error in AddQuote: %v", err)
		}
	}
}

func TestDeleteQuote(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}

	for i := 0; i < 10; i++ {
		quote := UniqueSpecifiedQuoteRequest{
			QuoteID: rand.Intn(20) - 10,
		}
		query := "UPDATE Quotes SET isAvailable=false, deletedAt=CURRENT_TIMESTAMP WHERE quoteID=?"
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID).WillReturnResult(sqlmock.NewResult(0, 1))

		err := w.DeleteQuote(quote)
		if err != nil {
			t.Errorf("Error in DeleteQuote: %v", err)
		}
	}
}

func TestGetQuotes(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := MultipleSpecifiedQuotesRequest{
		QuoteIDs: []string{"1", "2", "5"},
	}
	args := make([]interface{}, len(request.QuoteIDs))
	for i, quoteId := range request.QuoteIDs {
		args[i] = quoteId
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < len(request.QuoteIDs); i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}
	query := "SELECT .*? FROM Quotes LEFT JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true AND Quotes.quoteID IN \\(.*?\\) GROUP BY Quotes.quoteID"
	mock.ExpectQuery(query).WithArgs(args[0], args[1], args[2]).WillReturnRows(rows)

	_, err := w.GetQuotes(request)
	if err != nil {
		t.Errorf("Error in GetQuotes: %v", err)
	}
}

func TestGetLastQuotes(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := MultipleUnspecifiedQuotesRequest{
		QuoteNb: 5,
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < request.QuoteNb; i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}

	query := "SELECT .*? FROM Quotes LEFT OUTER JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID ORDER BY Quotes.quoteID DESC LIMIT .*? "
	mock.ExpectQuery(query).WithArgs(request.QuoteNb).WillReturnRows(rows)

	_, err := w.GetLastQuotes(request)
	if err != nil {
		t.Errorf("Error in GetLastQuotes: %v", err)
	}
}

func TestGetRandomQuotes(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := MultipleUnspecifiedQuotesRequest{
		QuoteNb: 5,
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < request.QuoteNb; i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}

	query := "SELECT .*? FROM Quotes LEFT JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID ORDER BY RANDOM\\(\\) LIMIT .*? "
	mock.ExpectQuery(query).WithArgs(request.QuoteNb).WillReturnRows(rows)

	_, err := w.GetRandomQuotes(request)
	if err != nil {
		t.Errorf("Error in GetRandomQuotes: %v", err)
	}
}

func TestGetTopQuotes(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := MultipleUnspecifiedQuotesRequest{
		QuoteNb: 5,
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < request.QuoteNb; i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}

	query := "SELECT .*? FROM Quotes JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID HAVING SUM\\(Votes.value\\) >= 0 ORDER BY SUM\\(Votes.value\\) DESC LIMIT .*? "

	mock.ExpectQuery(query).WithArgs(request.QuoteNb).WillReturnRows(rows)

	_, err := w.GetTopQuotes(request)
	if err != nil {
		t.Errorf("Error in GetTopQuotes: %v", err)
	}
}

func TestGetFlopQuotes(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}
	request := MultipleUnspecifiedQuotesRequest{
		QuoteNb: 5,
	}
	rows := sqlmock.NewRows([]string{"quoteID", "content", "context", "author", "CreatedAt", "DeletedAt", "IsActive", "Votes"})
	for i := 0; i < request.QuoteNb; i++ {
		u := QuoteResponse{
			QuoteID:      i,
			Author:       "a",
			Content:      "b",
			QuoteContext: "c",
			CreatedAt:    time.Time{},
			DeletedAt:    time.Time{},
			IsActive:     true,
			Votes:        1,
		}
		rows = rows.AddRow(u.QuoteID, u.Content, u.QuoteContext, u.Author, u.CreatedAt, u.DeletedAt, u.IsActive, u.Votes)
	}

	query := "SELECT .*? FROM Quotes JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID HAVING SUM\\(Votes.value\\) <= 0 ORDER BY SUM\\(Votes.value\\) ASC LIMIT .*? "
	mock.ExpectQuery(query).WithArgs(request.QuoteNb).WillReturnRows(rows)

	_, err := w.GetFlopQuotes(request)
	if err != nil {
		t.Errorf("Error in GetFlopQuotes: %v", err)
	}
}

func TestUnVoteQuote(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}

	for i := 0; i < 10; i++ {
		quote := VoteQuoteRequest{
			QuoteID: rand.Intn(10),
			Voter:   0,
		}
		query := "DELETE FROM Votes WHERE quoteID=.*? AND voter=.*?"
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID, quote.Voter).WillReturnResult(sqlmock.NewResult(0, 1))

		err := w.UnVoteQuote(quote)
		if err != nil {
			t.Errorf("Error in UnVote: %v", err)
		}
	}
}

func TestUpVoteQuote(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}

	for i := 0; i < 1; i++ {
		quote := VoteQuoteRequest{
			QuoteID: rand.Intn(10),
			Voter:   0,
		}

		query := "DELETE FROM Votes WHERE quoteID=.*? AND voter=.*?"
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID, quote.Voter).WillReturnResult(sqlmock.NewResult(0, 1))

		query = "INSERT INTO Votes \\(quoteID, voter, value\\) VALUES \\(.*?,.*?,1\\)"
		prep = mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID, quote.Voter).WillReturnResult(sqlmock.NewResult(0, 1))

		err := w.UpVoteQuote(quote)
		if err != nil {
			t.Errorf("Error in UpVote: %v", err)
		}
	}
}

func TestDownVoteQuote(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	w := SqliteWrapper{
		DB: db,
	}

	for i := 0; i < 1; i++ {
		quote := VoteQuoteRequest{
			QuoteID: rand.Intn(10),
			Voter:   0,
		}

		query := "DELETE FROM Votes WHERE quoteID=.*? AND voter=.*?"
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID, quote.Voter).WillReturnResult(sqlmock.NewResult(0, 1))

		query = "INSERT INTO Votes \\(quoteID, voter, value\\) VALUES \\(.*?,.*?,-1\\)"
		prep = mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(quote.QuoteID, quote.Voter).WillReturnResult(sqlmock.NewResult(0, 1))

		err := w.DownVoteQuote(quote)
		if err != nil {
			t.Errorf("Error in UpVote: %v", err)
		}
	}
}

func TestCheckContext(t *testing.T) {
	samples := []struct {
		Input  string
		Output bool
	}{
		{
			Input:  "Anonyme",
			Output: false,
		},
		{
			Input:  "un contexte normal",
			Output: true,
		},
		{
			Input:  "une personne anonyme",
			Output: true, // à changer si on met une regex
		},
	}

	for _, sample := range samples {
		isContextAllowed := checkContext(sample.Input)
		if isContextAllowed != sample.Output {
			t.Errorf("error during measureIndex for %s", sample.Input)
		}
	}
}
