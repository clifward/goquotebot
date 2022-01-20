package telegram

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"goquotebot/pkg/storages"
	"io/fs"
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

	templates map[string]*template.Template
	//go:embed templates/*
	files embed.FS
)

func init() {

	regexAddQuote = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(\S.+?)\s{0,}\|\s{0,}(\S.+?)\s{0,}$`)
	regexQuotesIDs = regexp.MustCompile(`(^|\s)#Q{0,}([0-9]{1,})\b`)
	regexCmdNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}#{0,}Q{0,}([0-9]{1,})\s{0,}$`)
	//regexSearchExpressionNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(.+?)\s{0,}(\d{0,})\s{0,}$`)
	regexSearchExpressionNumber = regexp.MustCompile(`^\/[a-zA-Z]{1,}\s{1,}(.+?)(\s{1,}(\d{1,})|$)`)

	templates = make(map[string]*template.Template)
	err := LoadTemplates("templates", "/*.tmpl")
	if err != nil {
		panic(err)
	}

}

func LoadTemplates(templatesDir string, extension string) error {
	tmplFiles, err := fs.ReadDir(files, "templates")
	if err != nil {
		fmt.Println("can't read templates dir")
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt

	}

	return err
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
	err := templates["quotes.tmpl"].Execute(&buf, quotes)
	if err != nil {
		return "", err
	}
	response := buf.String()
	return response[:len(response)-38], nil
}

func GenerateNewQuoteMessage(quote storages.AddQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templates["quote_added.tmpl"].Execute(&buf, quote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateDeleteQuoteMessage(quote storages.UniqueSpecifiedQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templates["quote_deleted.tmpl"].Execute(&buf, quote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateVoteAddedMessage(vote storages.VoteQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templates["vote_added.tmpl"].Execute(&buf, vote)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateVoteRemovedMessage(vote storages.VoteQuoteRequest) (string, error) {
	var buf bytes.Buffer
	err := templates["vote_removed.tmpl"].Execute(&buf, vote)
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
