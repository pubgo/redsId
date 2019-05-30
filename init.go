package redsid

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.NewConsoleWriter()).With().Str("module", "redsid").Caller().Logger()
}
