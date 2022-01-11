package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"goquotebot/pkg/storages"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	ErrNoIDProvided = errors.New("no id provided")

	regexAddQuote               *regexp.Regexp
	regexQuotesIDs              *regexp.Regexp
	regexCmdNumber              *regexp.Regexp
	regexSearchExpressionNumber *regexp.Regexp

	templateQuoteDeleted *template.Template
	templateVoteAdded    *template.Template
	templateVoteRemoved  *template.Template
	templateQuoteAdded   *template.Template
	templateQuotes       *template.Template
)

func init() {
	templatePath, err := localizeTemplate()
	if err != nil {
		panic(err)
	}

	regexAddQuote = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(\S.+?)\s{0,}\|\s{0,}(\S.+?)\s{0,}$`)
	regexQuotesIDs = regexp.MustCompile(`(^|\s)#Q{0,}([0-9]{1,})\b`)
	regexCmdNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}#{0,}Q{0,}([0-9]{1,})\s{0,}$`)
	//regexSearchExpressionNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(.+?)\s{0,}(\d{0,})\s{0,}$`)
	regexSearchExpressionNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(.+?)(\s{1,}(\d{1,})|$)`)

	templateQuoteDeleted = template.Must(template.New("quote_deleted.tmpl").ParseFiles(fmt.Sprintf("%squote_deleted.tmpl", templatePath)))
	templateVoteAdded = template.Must(template.New("vote_added.tmpl").ParseFiles(fmt.Sprintf("%svote_added.tmpl", templatePath)))
	templateVoteRemoved = template.Must(template.New("vote_removed.tmpl").ParseFiles(fmt.Sprintf("%svote_removed.tmpl", templatePath)))
	templateQuoteAdded = template.Must(template.New("quote_added.tmpl").ParseFiles(fmt.Sprintf("%squote_added.tmpl", templatePath)))
	templateQuotes = template.Must(template.New("quotes.tmpl").ParseFiles(fmt.Sprintf("%squotes.tmpl", templatePath)))

}

func localizeTemplate() (string, error) {
	pathsToTry := []string{"pkg/telegram/templates/", "../../pkg/telegram/templates/"}
	for _, path := range pathsToTry {
		file := fmt.Sprintf("%squote_deleted.tmpl", path)
		if fileExists(file) {
			return path, nil
		}
	}
	return "", errors.New("templates not found")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ExtractExpressionAndNumber(t string) storages.SearchExpressionRequest {
	matches := regexSearchExpressionNumber.FindAllStringSubmatch(t, -1)
	if len(matches) == 0 {
		return storages.SearchExpressionRequest{}
	}

	match := matches[0]
	if len(match) < 2 {
		return storages.SearchExpressionRequest{}
	}
	if len(match) == 2 {
		return storages.SearchExpressionRequest{
			Expression: strings.TrimSpace(match[1]),
			QuoteNb:    1,
		}
	}

	nb, err := strconv.Atoi(match[3])
	if err != nil {
		nb = 1
	}

	return storages.SearchExpressionRequest{
		Expression: strings.TrimSpace(match[1]),
		QuoteNb:    nb,
	}
}

func ExtractQuotesID(t string) []string {
	matches := regexQuotesIDs.FindAllStringSubmatch(t, -1)
	res := make([]string, 0)
	for _, match := range matches {
		res = append(res, match[2])
	}
	return res
}

func ExtractQuote(t string) []string {
	matches := regexAddQuote.FindAllStringSubmatch(t, -1)
	res := make([]string, 0)
	for _, match := range matches {
		res = append(res, match[1])
		res = append(res, match[2])
	}
	return res
}

func ConvertMatchToInt(m []string) (int, error) {
	res, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}

	return res, nil
}

func ExtractID(t string) (int, error) {
	matches := regexCmdNumber.FindAllStringSubmatch(t, -1)
	if len(matches) < 1 {
		return 0, ErrNoIDProvided
	}

	return ConvertMatchToInt(matches[0])
}

func ExtractNumber(t string) (int, error) {
	matches := regexCmdNumber.FindAllStringSubmatch(t, -1)
	if len(matches) < 1 {
		return 1, nil
	}

	return ConvertMatchToInt(matches[0])
}

func GenerateQuotesMessage(quotes []storages.QuoteResponse) (string, error) {
	if len(quotes) == 0 {
		return "No quote available", nil
	}

	var buf bytes.Buffer
	err := templateQuotes.Execute(&buf, quotes)
	if err != nil {
		return "", err
	}
	response := buf.String()
	return response[:len(response)-38], nil
}

func GenerateNewQuoteMessage(quote storages.AddQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templateQuoteAdded.Execute(&buf, quote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateDeleteQuoteMessage(quote storages.UniqueSpecifiedQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templateQuoteDeleted.Execute(&buf, quote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateVoteAddedMessage(vote storages.VoteQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templateVoteAdded.Execute(&buf, vote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateVoteRemovedMessage(vote storages.VoteQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templateVoteRemoved.Execute(&buf, vote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func isAtLeastMember(member *tb.ChatMember) bool {
	switch member.Role {
	case "member":
		return true
	case "administrator":
		return true
	case "creator":
		return true
	default:
		return false
	}
}

func isAtLeastAdmin(member *tb.ChatMember) bool {
	switch member.Role {
	case "administrator":
		return true
	case "creator":
		return true
	default:
		return false
	}
}
