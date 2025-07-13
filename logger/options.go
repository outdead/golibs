package logger

import (
	"github.com/bwmarrin/discordgo"
)

type Option func(log *Logger)

func WithDiscordSession(session *discordgo.Session) Option {
	return func(log *Logger) {
		log.discordSession = session
	}
}
