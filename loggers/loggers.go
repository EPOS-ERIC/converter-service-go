package loggers

import (
	"log"
	"os"
)

var EA_LOGGER = log.New(os.Stdout, "[External Access] ", log.LstdFlags)
var PS_LOGGER = log.New(os.Stdout, "[Processing Service] ", log.LstdFlags)
var RS_LOGGER = log.New(os.Stdout, "[Resources Service] ", log.LstdFlags)
var API_LOGGER = log.New(os.Stdout, "[API] ", log.LstdFlags)
