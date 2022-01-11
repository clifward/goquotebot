package telegram

import (
	"errors"
	"fmt"
	c "goquotebot/pkg/storages"

	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (s *Server) Message(m *tb.Message) (*tb.Message, error) {
	IDs := ExtractQuotesID(m.Text)
	if len(IDs) == 0 {
		return nil, errors.New("no id provided")
	}

	quotes, err := (*s.DB).GetQuotes(c.MultipleSpecifiedQuotesRequest{QuoteIDs: IDs})
	if err != nil {
		s.Logger.Error("failed to fetch quotes by ids", zap.Error(err), zap.Strings("IDs", IDs))
		return nil, err
	}

	if len(quotes) == 0 {
		return nil, nil
	}

	response, err := GenerateQuotesMessage(quotes)
	if err != nil {
		s.Logger.Error("failed to generate quote message", zap.Error(err), zap.Any("quotes", quotes))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) AddQuote(m *tb.Message) (*tb.Message, error) {
	var err error
	if m.FromGroup() {
		err = s.Bot.Delete(m)
		if err != nil {
			s.Logger.Error("failed to delete a message", zap.Error(err), zap.Any("message to delete", m))
		}
	}

	tmp := ExtractQuote(m.Text)
	if len(tmp) < 2 {
		s.Bot.Send(m.Sender, "Cannot add this quote")
		return nil, errors.New("not enough argument provided")
	}

	quote := c.AddQuoteRequest{
		Author:       m.Sender.Username,
		Content:      tmp[0],
		QuoteContext: tmp[1],
	}

	message, err := (*s.DB).AddQuote(quote)
	if err != nil {
		s.Logger.Error("failed to add a quote", zap.Error(err), zap.Any("quote", quote))
		return nil, errors.New("cannot add the quote to the DB")
	}

	if message != "" {
		s.Bot.Send(m.Sender, message)
		senderChat, _ := s.Bot.ChatByID(fmt.Sprint(m.Sender.ID))
		s.Message(&tb.Message{Sender: m.Sender, Chat: senderChat, Text: message})
		return nil, nil
	}

	response, err := GenerateNewQuoteMessage(quote)
	if err != nil {
		s.Logger.Error("failed to generate quote message", zap.Error(err), zap.Any("quote", quote))
		return nil, err
	}

	// Sent the quote to the group chat
	s.Bot.Send(s.Chat, response)

	return s.Bot.Send(m.Sender, response)
}

func (s *Server) RandomQuotes(m *tb.Message) (*tb.Message, error) {
	res, err := ExtractNumber(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract number from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}

	quoteResponses, err := (*s.DB).GetRandomQuotes(c.MultipleUnspecifiedQuotesRequest{QuoteNb: res})
	if err != nil {
		s.Logger.Error("failed to get random quotes", zap.Error(err), zap.Int("QuoteNb", res))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) LastQuotes(m *tb.Message) (*tb.Message, error) {
	res, err := ExtractNumber(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract number from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	quoteResponses, err := (*s.DB).GetLastQuotes(c.MultipleUnspecifiedQuotesRequest{QuoteNb: res})
	if err != nil {
		s.Logger.Error("failed to get last quotes", zap.Error(err), zap.Int("QuoteNb", res))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) DeleteQuote(m *tb.Message) (*tb.Message, error) {
	var err error
	if m.FromGroup() {
		err = s.Bot.Delete(m)
		if err != nil {
			s.Logger.Error("failed to delete a message", zap.Error(err), zap.Any("message to delete", m))
		}
	}

	res, err := ExtractID(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract ID from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}

	err = (*s.DB).DeleteQuote(c.UniqueSpecifiedQuoteRequest{QuoteID: res})
	if err != nil {
		s.Logger.Error("failed to delete quote", zap.Error(err), zap.Int("QuoteID", res))
		return nil, err
	}

	response, err := GenerateDeleteQuoteMessage(c.UniqueSpecifiedQuoteRequest{QuoteID: res})
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Int("QuoteID", res))
		return nil, err
	}

	return s.Bot.Send(m.Sender, response)
}

func (s *Server) UpVote(m *tb.Message) (*tb.Message, error) {
	var err error
	if m.FromGroup() {
		err = s.Bot.Delete(m)
		if err != nil {
			s.Logger.Error("failed to delete a message", zap.Error(err), zap.Any("message to delete", m))
		}
	}

	res, err := ExtractID(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract ID from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	request := c.VoteQuoteRequest{
		QuoteID: res,
		Voter:   m.Sender.ID,
	}
	err = (*s.DB).UpVoteQuote(request)
	if err != nil {
		s.Logger.Error("failed to up vote a quote", zap.Error(err), zap.Any("vote quote request", request))
		return nil, err
	}

	response, err := GenerateVoteAddedMessage(request)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("request", request))
		return nil, err
	}

	return s.Bot.Send(m.Sender, response)
}

func (s *Server) DownVote(m *tb.Message) (*tb.Message, error) {
	var err error
	if m.FromGroup() {
		err = s.Bot.Delete(m)
		if err != nil {
			s.Logger.Error("failed to delete a message", zap.Error(err), zap.Any("message to delete", m))
		}
	}

	res, err := ExtractID(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract ID from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	request := c.VoteQuoteRequest{
		QuoteID: res,
		Voter:   m.Sender.ID,
	}
	err = (*s.DB).DownVoteQuote(request)
	if err != nil {
		s.Logger.Error("failed to down vote a quote", zap.Error(err), zap.Any("vote quote request", request))
		return nil, err
	}

	response, err := GenerateVoteAddedMessage(request)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("request", request))
		return nil, err
	}

	return s.Bot.Send(m.Sender, response)
}

func (s *Server) UnVote(m *tb.Message) (*tb.Message, error) {
	var err error
	if m.FromGroup() {
		err = s.Bot.Delete(m)
		if err != nil {
			s.Logger.Error("failed to delete a message", zap.Error(err), zap.Any("message to delete", m))
		}
	}

	res, err := ExtractID(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract ID from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	request := c.VoteQuoteRequest{
		QuoteID: res,
		Voter:   m.Sender.ID,
	}
	err = (*s.DB).UnVoteQuote(request)
	if err != nil {
		s.Logger.Error("failed to unvote a quote", zap.Error(err), zap.Any("vote quote request", request))
		return nil, err
	}

	response, err := GenerateVoteRemovedMessage(request)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("request", request))
		return nil, err
	}

	return s.Bot.Send(m.Sender, response)
}

func (s *Server) TopQuotes(m *tb.Message) (*tb.Message, error) {
	res, err := ExtractNumber(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract number from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	quoteResponses, err := (*s.DB).GetTopQuotes(c.MultipleUnspecifiedQuotesRequest{QuoteNb: res})
	if err != nil {
		s.Logger.Error("failed to get top ranking", zap.Error(err), zap.Int("QuoteNb", res))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) FlopQuotes(m *tb.Message) (*tb.Message, error) {
	res, err := ExtractNumber(m.Text)
	if err != nil {
		s.Logger.Error("failed to extract number from command", zap.Error(err), zap.String("text", m.Text))
		return nil, err
	}
	quoteResponses, err := (*s.DB).GetFlopQuotes(c.MultipleUnspecifiedQuotesRequest{QuoteNb: res})
	if err != nil {
		s.Logger.Error("failed to get flop ranking", zap.Error(err), zap.Int("QuoteNb", res))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) SearchQuote(m *tb.Message) (*tb.Message, error) {
	res := ExtractExpressionAndNumber(m.Text)

	quoteResponses, err := (*s.DB).SearchExpression(res)
	if err != nil {
		s.Logger.Error("failed to get flop ranking", zap.Error(err), zap.Int("QuoteNb", res.QuoteNb))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}

func (s *Server) SearchWordQuote(m *tb.Message) (*tb.Message, error) {
	res := ExtractExpressionAndNumber(m.Text)

	quoteResponses, err := (*s.DB).SearchWord(res)
	if err != nil {
		s.Logger.Error("failed to search word", zap.Error(err), zap.Any("QuoteNb", res))
		return nil, err
	}

	response, err := GenerateQuotesMessage(quoteResponses)
	if err != nil {
		s.Logger.Error("failed to generate quotes message", zap.Error(err), zap.Any("quotes", quoteResponses))
		return nil, err
	}

	return s.Bot.Send(m.Chat, response)
}
