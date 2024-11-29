package loggers

import (
	"log"
	"os"
)

var EA_LOGGER = log.New(os.Stdout, "[External Access] ", log.LstdFlags)
var PS_LOGGER = log.New(os.Stdout, "[Processing Service] ", log.LstdFlags)
var SERVICE_LOGGER = log.New(os.Stdout, "[Service] ", log.LstdFlags)
