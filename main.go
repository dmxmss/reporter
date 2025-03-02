package main

import (
  "github.com/dmxmss/reporter/src/bot"
  "log"
)

func main() {
  githubClient := github_client.newGitHubClient()

  if bot, err := bot.initBot(); err != nil {
    log.Fatal(err)
    return
  }
  
  bot.Debug = true

  log.Printf("Authorized on account %s", bot.Self.UserName)

  bot.startPolling()
}
