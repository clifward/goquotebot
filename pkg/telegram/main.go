package telegram

import (
	"errors"
	"fmt"
	"goquotebot/internal/monitoring/metrics"
	"goquotebot/pkg/config"
	"os"
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

	path, err := localizeDB()
	if err != nil {
		return nil, err
	}
	pathDB := fmt.Sprintf("%squotes.db", path)
	db, err := c.NewSqliteWrapper(pathDB)
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

func localizeDB() (string, error) {
	pathsToTry := []string{"sqlite/", "../../cmd/goquote/"}
	for _, path := range pathsToTry {
		if checkIfFolderExists(path) {
			return path, nil
		}
	}
	return "", errors.New("DB folder not found")
}

func checkIfFolderExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
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
