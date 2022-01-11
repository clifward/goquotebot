package storages

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	_ "github.com/mattn/go-sqlite3"
)

type Pair struct {
	Key   int
	Value float64
}

type PairList []Pair

type SqliteWrapper struct {
	DB *sql.DB
}

func NewSqliteWrapper(pathDB string) (DB, error) {
	if !fileExists(pathDB) {
		fmt.Println("db does not exist, attempting to create it")
		file, err := os.Create(pathDB) // Create SQLite file
		if err != nil {
			return nil, err
		}
		file.Close()
	}
	db, err := sql.Open("sqlite3", pathDB)
	if err != nil {
		return nil, err
	}
	wrapper := SqliteWrapper{
		DB: db,
	}
	err = wrapper.CreateQuotesTable()
	if err != nil {
		return nil, err
	}
	err = wrapper.CreateVotesTable()
	if err != nil {
		return nil, err
	}
	return &wrapper, nil
}

func (w *SqliteWrapper) Close() error {
	return w.DB.Close()
}

//to authorize foreing keys (if needed) : PRAGMA foreign_keys = ON;
func (w *SqliteWrapper) CreateQuotesTable() error {
	_, err := w.DB.Exec("CREATE TABLE IF NOT EXISTS Quotes (quoteID INTEGER PRIMARY KEY AUTOINCREMENT, `content` VARCHAR(512) NOT NULL, `context` VARCHAR(255) NOT NULL, `author` VARCHAR(255) NOT NULL, `createdAt` DATETIME DEFAULT CURRENT_TIMESTAMP, `deletedAt` DATETIME DEFAULT NULL, `isAvailable` BOOLEAN NOT NULL) ; UPDATE SQLITE_SEQUENCE SET seq=100 WHERE name='Quotes'")
	return err
}

func (w *SqliteWrapper) CreateVotesTable() error {
	_, err := w.DB.Exec("CREATE TABLE IF NOT EXISTS Votes (`voteID` INTEGER PRIMARY KEY AUTOINCREMENT, `quoteID` INTEGER NOT NULL, `voter` sqlite3_int64 NOT NULL, `value` INTEGER NOT NULL, FOREIGN KEY(`quoteID`) REFERENCES Quotes(`quoteID`))")
	return err
}

func (w *SqliteWrapper) DeleteQuotesTable() error {
	_, err := w.DB.Exec("DROP TABLE IF EXISTS Quotes;")
	return err
}

func (w *SqliteWrapper) DeleteVotesTable() error {
	_, err := w.DB.Exec("DROP TABLE IF EXISTS Votes;")
	return err
}

func (w *SqliteWrapper) AddQuote(request AddQuoteRequest) (string, error) {
	contextIsAllowed := checkContext(request.QuoteContext)
	if !contextIsAllowed {
		message := fmt.Sprintf("🚫 Quote not added 🚫 \n Your context:\n *%s* \n\nis forbidden \n", request.QuoteContext)
		return message, nil
	}

	isProbablyStored, _, quoteIdOfMax, err := w.checkIfExists(request)
	if err != nil {
		return "", err
	}
	if isProbablyStored {
		message := fmt.Sprintf("🚫 Quote not added 🚫 \n Your quote:\n *%s* \n\nis very similar to quote #Q%d \n", request.Content, quoteIdOfMax)
		return message, nil
	}

	query := "INSERT INTO Quotes (content, context, author, createdAt, isAvailable) VALUES (?,?,?,CURRENT_TIMESTAMP,?)"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt, err := w.DB.PrepareContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, request.Content, request.QuoteContext, request.Author, 1)
	return "", err
}

func (w *SqliteWrapper) DeleteQuote(request UniqueSpecifiedQuoteRequest) error {
	query := "UPDATE Quotes SET isAvailable=false, deletedAt=CURRENT_TIMESTAMP WHERE quoteID=?"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt, err := w.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, request.QuoteID)
	return err
}

func (w *SqliteWrapper) GetQuotes(request MultipleSpecifiedQuotesRequest) ([]QuoteResponse, error) {
	if len(request.QuoteIDs) == 0 {
		return []QuoteResponse{}, nil
	}

	args := make([]interface{}, len(request.QuoteIDs))
	for i, quoteId := range request.QuoteIDs {
		args[i] = quoteId
	}
	query := "SELECT Quotes.*,SUM(Votes.value) FROM Quotes LEFT JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true AND Quotes.quoteID IN ( ?" + strings.Repeat(",?", len(args)-1) + " ) GROUP BY Quotes.quoteID"

	var value []QuoteResponse
	results, err := w.DB.Query(query, args...)
	if err != nil {
		return value, err
	}

	defer results.Close()
	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) GetLastQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error) {
	query := "SELECT Quotes.*, SUM(Votes.value) FROM Quotes LEFT OUTER JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID ORDER BY Quotes.quoteID DESC LIMIT ? "
	results, err := w.DB.Query(query, request.QuoteNb)
	if err != nil {
		return nil, err
	}
	defer results.Close()
	var value []QuoteResponse
	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) GetRandomQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error) {
	query := "SELECT Quotes.*,SUM(Votes.value) FROM Quotes LEFT JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID ORDER BY RANDOM() LIMIT ? "
	var value []QuoteResponse
	results, err := w.DB.Query(query, request.QuoteNb)
	if err != nil {
		return value, err
	}
	defer results.Close()

	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) GetTopQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error) {
	query := "SELECT Quotes.*,SUM(Votes.value) FROM Quotes JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID HAVING SUM(Votes.value) >= 0 ORDER BY SUM(Votes.value) DESC LIMIT ? "
	var value []QuoteResponse
	results, err := w.DB.Query(query, request.QuoteNb)
	if err != nil {
		return value, err
	}
	defer results.Close()

	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) GetFlopQuotes(request MultipleUnspecifiedQuotesRequest) ([]QuoteResponse, error) {
	query := "SELECT Quotes.*,SUM(Votes.value) FROM Quotes JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true GROUP BY Quotes.quoteID HAVING SUM(Votes.value) <= 0 ORDER BY SUM(Votes.value) ASC LIMIT ? "
	var value []QuoteResponse
	results, err := w.DB.Query(query, request.QuoteNb)
	if err != nil {
		return value, err
	}
	defer results.Close()

	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) UnVoteQuote(request VoteQuoteRequest) error {
	query := "DELETE FROM Votes WHERE quoteID=? AND voter=?"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt, err := w.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, request.QuoteID, request.Voter)
	return err
}

func (w *SqliteWrapper) UpVoteQuote(request VoteQuoteRequest) error {
	err := w.UnVoteQuote(request)
	if err != nil {
		return err
	}
	query := "INSERT INTO Votes (quoteID, voter, value) VALUES (?,?,1)"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt, err := w.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, request.QuoteID, request.Voter)
	return err
}

func (w *SqliteWrapper) DownVoteQuote(request VoteQuoteRequest) error {
	err := w.UnVoteQuote(request)
	if err != nil {
		return err
	}
	query := "INSERT INTO Votes (quoteID, voter, value) VALUES (?,?,-1)"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt, err := w.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, request.QuoteID, request.Voter)
	return err
}

func (w *SqliteWrapper) SearchWord(request SearchExpressionRequest) ([]QuoteResponse, error) {
	query := `SELECT Quotes.*,SUM(Votes.value) FROM Quotes LEFT JOIN Votes ON Quotes.quoteID = Votes.quoteID WHERE Quotes.isAvailable=true AND Quotes.content LIKE ? GROUP BY Quotes.quoteID ORDER BY RANDOM() LIMIT ?`
	var value []QuoteResponse
	results, err := w.DB.Query(query, "%"+request.Expression+"%", request.QuoteNb)
	if err != nil {
		return value, err
	}
	defer results.Close()

	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			panic(err.Error())
		}
		value = append(value, quote)
	}
	return value, err
}

func (w *SqliteWrapper) SearchExpression(request SearchExpressionRequest) ([]QuoteResponse, error) {
	quoteArray, err := w.getAllDBContent()
	if err != nil {
		return []QuoteResponse{}, err
	}
	threshold := 0.1
	_, similarQuotes, _ := measureIndex(request.Expression, quoteArray, threshold)
	topMatchQuotes := topMatch(similarQuotes, request.QuoteNb, threshold)
	resultQuoteIDs := make([]string, 0)
	for _, quoteID := range topMatchQuotes {
		resultQuoteIDs = append(resultQuoteIDs, fmt.Sprint(quoteID))
	}
	fmt.Println("quoteId", resultQuoteIDs)
	results, err := w.GetQuotes(MultipleSpecifiedQuotesRequest{
		QuoteIDs: resultQuoteIDs,
	})
	return results, err
}

//============================
//helpers, appendice functions

func checkContext(context string) bool {
	blacklist := []string{"Anonyme", "anonyme", "Anonymous"} //en attendant de faire une regex
	for _, forbiddenWord := range blacklist {
		if context == forbiddenWord {
			return false
		}
	}
	return true
}

func topMatch(similarQuotes map[int]float64, max int, threshold float64) []int {
	tops := []int{}
	p := make(PairList, len(similarQuotes))

	i := 0
	for k, v := range similarQuotes {
		p[i] = Pair{k, v}
		i++
	}
	sort.SliceStable(p, func(i, j int) bool {
		return p[i].Value > p[j].Value
	})
	for _, k := range p {
		if k.Value > threshold {
			tops = append(tops, k.Key)
		}
	}
	return tops
}

func (w *SqliteWrapper) getAllDBContent() (map[int]string, error) {
	quoteArray := make(map[int]string, 0)
	query := "SELECT Quotes.quoteID,content FROM Quotes WHERE isAvailable=true "
	results, err := w.DB.Query(query)
	if err != nil {
		return quoteArray, err
	}
	defer results.Close()
	for results.Next() {
		var quote QuoteResponse
		var quoteID sql.NullInt32
		var content sql.NullString
		err := results.Scan(&quoteID, &content)
		if err != nil {
			return quoteArray, err
		}
		if !quoteID.Valid {
			return quoteArray, nil
		} else {
			quote = QuoteResponse{
				QuoteID: int(quoteID.Int32),
				Content: content.String,
			}
		}
		if err != nil {
			return quoteArray, err
		}
		quoteArray[quote.QuoteID] = quote.Content
	}
	return quoteArray, nil
}

func (w *SqliteWrapper) checkIfExists(request AddQuoteRequest) (bool, map[int]float64, int, error) {
	similarQuotes := make(map[int]float64, 5)
	quoteArray, err := w.getLast5Contents()
	if err != nil {
		return false, similarQuotes, 0, err
	}
	//check if stored in the 5 last quotes:
	isStored, similarQuotes, max := measureIndex(request.Content, quoteArray, 0.45)
	return isStored, similarQuotes, max, nil
}

func (w *SqliteWrapper) getLast5Contents() (map[int]string, error) {
	query := "SELECT *,0 FROM Quotes WHERE Quotes.isAvailable=true ORDER BY Quotes.quoteID DESC LIMIT 5 "
	results, err := w.DB.Query(query)
	quoteArray := make(map[int]string, 0)
	if err != nil {
		return quoteArray, err
	}
	defer results.Close()
	for results.Next() {
		quote, err := ScanFromResults(results)
		if err != nil {
			return quoteArray, err
		}
		quoteArray[quote.QuoteID] = quote.Content
	}
	return quoteArray, nil

}

func measureIndex(quote string, quoteArray map[int]string, threshold float64) (bool, map[int]float64, int) {
	results := make(map[int]float64, 5)
	isProbablyStored := false
	QuoteIdOfMax, val := 0, 0.0
	for quoteID, storedQuote := range quoteArray {
		sd := metrics.NewSorensenDice()
		similarity := strutil.Similarity(quote, storedQuote, sd)
		results[quoteID] = similarity
		if similarity >= threshold {
			results[quoteID] = similarity
			isProbablyStored = true
			if similarity > val {
				val = similarity
				QuoteIdOfMax = quoteID
			}
		}
	}
	return isProbablyStored, results, QuoteIdOfMax
}

func sqliteTsToTime(val sql.NullString) (time.Time, error) {
	if !val.Valid {
		return time.Time{}, nil
	} else {
		t, err := time.Parse("2006-01-02T15:04:05Z", val.String)
		return t, err
	}

}

func ScanFromResults(results *sql.Rows) (QuoteResponse, error) {
	var quote QuoteResponse
	var deletionDate sql.NullString
	var creationDate sql.NullString
	var votes sql.NullInt32
	var quoteID sql.NullInt32
	var content sql.NullString
	var quoteContext sql.NullString
	var author sql.NullString
	var isActive sql.NullBool
	err := results.Scan(&quoteID, &content, &quoteContext, &author, &creationDate, &deletionDate, &isActive, &votes)
	if err != nil {
		return quote, err
	}
	if !quoteID.Valid {
		return QuoteResponse{}, nil
	} else {
		quote = QuoteResponse{
			QuoteID:      int(quoteID.Int32),
			Content:      content.String,
			QuoteContext: quoteContext.String,
			Author:       author.String,
			IsActive:     isActive.Bool,
		}
		quote.CreatedAt, err = sqliteTsToTime(creationDate)
		if err != nil {
			return quote, err
		}
		quote.DeletedAt, err = sqliteTsToTime(deletionDate)
		if err != nil {
			return quote, err
		}
		if !votes.Valid {
			quote.Votes = 0
		} else {
			quote.Votes = int(votes.Int32)
		}
		return quote, err
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
