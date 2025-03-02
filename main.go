package main

import (
  "reporter/internal/bot"
  "log"
)

func main() {
  bot, err := bot.InitBot()

  if err != nil {
    log.Fatal(err)
    return
  }
  
  bot.Debug = true

  log.Printf("Authorized on account %s", bot.Self.UserName)

  bot.StartPolling()
}
