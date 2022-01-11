package telegram

import (
	"errors"
	"goquotebot/pkg/config"
	"testing"

	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	samples := []struct {
		Input            *config.Config
		ErrInitExpected  error
		ErrCloseExpected error
	}{
		{
			Input: &config.Config{
				Telegram: config.TelegramConfig{
					Token: "wrong token",
				},
			},
			ErrInitExpected:  errors.New("telegram: Not Found (404)"),
			ErrCloseExpected: nil,
		},
	}

	for _, sample := range samples {
		server, err := NewServer(logger, sample.Input)
		if !(err != nil && sample.ErrInitExpected != nil && err.Error() == sample.ErrInitExpected.Error()) {
			t.Errorf("got %v instead of %v", err, sample.ErrInitExpected)
		}

		if err == nil && sample.ErrInitExpected == nil {
			server.Start()
			err = server.Stop()
			if err != sample.ErrCloseExpected {
				t.Errorf("got %v instead of %v", err, sample.ErrCloseExpected)
			}
		}
	}
}
