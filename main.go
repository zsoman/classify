package main

import (
	"os"
	"./servers"
	log "github.com/Sirupsen/logrus"
	
)

func main() {
	log.Warn("Server is started.")
	server.StartServer()
}

func init() {
  log.SetFormatter(&log.TextFormatter{})
  log.SetOutput(os.Stderr)
  log.SetLevel(log.WarnLevel)
}