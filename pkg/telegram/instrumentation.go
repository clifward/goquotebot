package telegram

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	commandsReceived *prometheus.CounterVec
	messagesReceived prometheus.Counter
	commandsTriggers *prometheus.CounterVec
)

func init() {
	commandsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telegram_commands_received",
		Help: "The total number of commands received",
	}, []string{"command"})

	messagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "telegram_messages_received",
		Help: "The total number of messages received",
	})

	commandsTriggers = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_triggered_counter",
		Help: "The number of trigger of each command with the resulting status code",
	}, []string{"command", "status"})
}
