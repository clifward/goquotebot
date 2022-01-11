package telegram

import (
	"goquotebot/internal/monitoring/metrics"
	"goquotebot/pkg/config"
	"time"

	c "goquotebot/pkg/storages"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Server struct {
	Bot    *tb.Bot
	DB     *c.DB
	Chat   *tb.Chat
	Logger *zap.Logger
	ms     *metrics.MonitoringServer
}

func NewServer(logger *zap.Logger, cfg *config.Config) (*Server, error) {
	b, err := tb.NewBot(tb.Settings{
		Token:     cfg.Telegram.Token,
		Poller:    &tb.LongPoller{Timeout: 1 * time.Second},
		ParseMode: tb.ModeMarkdown,
	})
	if err != nil {
		return nil, err
	}

	chat, err := b.ChatByID(cfg.Telegram.GroupID)
	if err != nil {
		return nil, err
	}

	db, err := c.NewSqliteWrapper("../../cmd/goquote/quotes.db")
	if err != nil {
		return nil, err
	}

	server := &Server{
		Bot:    b,
		DB:     &db,
		Chat:   chat,
		Logger: logger,
	}

	err = server.RegisterRoutes()
	if err != nil {
		return nil, err
	}

	ms, err := metrics.StartMonitoringServer(logger, cfg.Metrics)
	if err != nil {
		return nil, err
	}

	server.ms = ms

	return server, nil
}

func (s *Server) Start() {
	s.Bot.Start()
}

func (s *Server) Stop() error {
	var errs error

	err := (*s.DB).Close()
	if err != nil {
		s.Logger.Error("failed to close the DB server")
		errs = multierror.Append(errs, err)
	}

	err = s.ms.Stop()
	if err != nil {
		s.Logger.Error("failed to close the monitoring server")
		errs = multierror.Append(errs, err)
	}

	s.Bot.Stop()

	return errs
}
