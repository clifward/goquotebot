package telegram

import (
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

func Anyone(s *Server, m *tb.Message, f func(*tb.Message) (*tb.Message, error)) func(*tb.Message) (*tb.Message, error) {
	return f
}

func MustBeMember(s *Server, m *tb.Message, f func(*tb.Message) (*tb.Message, error)) func(*tb.Message) (*tb.Message, error) {
	member, err := s.Bot.ChatMemberOf(s.Chat, m.Sender)
	if err != nil {
		return func(t *tb.Message) (*tb.Message, error) {
			s.Logger.Error("failed to check the status of a user", zap.Error(err), zap.Any("Chat", s.Chat), zap.Any("user", m.Sender))
			return nil, err
		}
	}

	if !isAtLeastMember(member) {
		return func(t *tb.Message) (*tb.Message, error) {
			s.Bot.Send(t.Sender, "You must be at least a registered member to do this.")
			s.Logger.Info("unauthorized user spoke to the bot", zap.Any("user", t.Sender), zap.String("message", t.Text))
			return nil, nil
		}
	}

	return f
}

func MustBeAdministrator(s *Server, m *tb.Message, f func(*tb.Message) (*tb.Message, error)) func(*tb.Message) (*tb.Message, error) {
	member, err := s.Bot.ChatMemberOf(s.Chat, m.Sender)
	if err != nil {
		return func(t *tb.Message) (*tb.Message, error) {
			s.Logger.Error("failed to check the status of a user", zap.Error(err), zap.Any("Chat", s.Chat), zap.Any("user", m.Sender))
			return nil, err
		}
	}

	if !isAtLeastAdmin(member) {
		return func(t *tb.Message) (*tb.Message, error) {
			s.Bot.Send(t.Sender, "You must be at least an administrator to do this.")
			s.Logger.Info("unauthorized user spoke to the bot", zap.Any("user", t.Sender), zap.String("message", t.Text))
			return nil, nil
		}
	}

	return f
}
