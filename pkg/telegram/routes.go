package telegram

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type SuperCommand struct {
	Command        tb.Command
	Handler        func(*tb.Message) (*tb.Message, error)
	AuthMiddleware func(*Server, *tb.Message, func(*tb.Message) (*tb.Message, error)) func(*tb.Message) (*tb.Message, error)
}

func (server *Server) RegisterRoutes() error {

	cmds := []SuperCommand{
		{
			Command: tb.Command{
				Text:        "add",
				Description: "Usage : /add quote | context",
			},
			Handler:        server.AddQuote,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "random",
				Description: "Usage : /random <n> with <n> the number of random quotes to fetch",
			},
			Handler:        server.RandomQuotes,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "last",
				Description: "Usage : /last <n> with <n> the number of last quotes to fetch",
			},
			Handler:        server.LastQuotes,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "delete",
				Description: "Usage : /delete <id> with <id> the ID of the quote",
			},
			Handler:        server.DeleteQuote,
			AuthMiddleware: MustBeAdministrator,
		},
		{
			Command: tb.Command{
				Text:        "upvote",
				Description: "Usage : /upvote <id> will add +1 vote to the <ID> quote",
			},
			Handler:        server.UpVote,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "downvote",
				Description: "Usage : /downvote <id> will add -1 vote to the <ID> quote",
			},
			Handler:        server.DownVote,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "unvote",
				Description: "Usage : /unvote <id> will remove your vote on the <ID> quote",
			},
			Handler:        server.UnVote,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "top",
				Description: "Usage : /top <n> will show the <n> most liked quotes",
			},
			Handler:        server.TopQuotes,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "flop",
				Description: "Usage : /flop <n> will show the <n> most disliked quotes",
			},
			Handler:        server.FlopQuotes,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "s",
				Description: "Usage : /s expression <n> will show at most <n> quotes close to <expression>",
			},
			Handler:        server.SearchQuote,
			AuthMiddleware: MustBeMember,
		},
		{
			Command: tb.Command{
				Text:        "sw",
				Description: "Usage : /sw <word> <n> will show at most <n> quotes containing the word <word>",
			},
			Handler:        server.SearchWordQuote,
			AuthMiddleware: MustBeMember,
		},
	}

	botCommands := make([]tb.Command, 0)
	for _, cmd := range cmds {
		h := cmd
		server.Bot.Handle(fmt.Sprintf("/%s", h.Command.Text), func(m *tb.Message) {
			server.Logger.Debug("command received", zap.String("command", m.Text), zap.Any("user", m.Sender), zap.Any("chat", m.Chat))
			commandsReceived.With(prometheus.Labels{"command": h.Command.Text}).Inc()

			content, err := h.AuthMiddleware(server, m, h.Handler)(m)
			if err != nil {
				if content != nil {
					server.Logger.Error("failed to send message", zap.Error(err), zap.Any("response", content))
				}
				commandsTriggers.With(prometheus.Labels{"command": h.Command.Text, "status": "424"}).Inc()
				return
			}
			commandsTriggers.With(prometheus.Labels{"command": h.Command.Text, "status": "200"}).Inc()
		})
		botCommands = append(botCommands, h.Command)
	}

	err := server.Bot.SetCommands(botCommands)
	if err != nil {
		return err
	}

	server.Bot.Handle(tb.OnText, func(m *tb.Message) {
		server.Logger.Debug("command received", zap.String("command", m.Text), zap.Any("user", m.Sender), zap.Any("chat", m.Chat))
		messagesReceived.Inc()

		content, err := MustBeMember(server, m, server.Message)(m)
		if err != nil {
			if content != nil {
				server.Logger.Error("failed to send message", zap.Error(err), zap.Any("response", content))
			}
			commandsTriggers.With(prometheus.Labels{"command": "message", "status": "424"}).Inc()
			return
		}
		commandsTriggers.With(prometheus.Labels{"command": "message", "status": "200"}).Inc()
	})

	return nil
}
