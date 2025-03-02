package bot

import (
  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
  "github.com/dmxmss/reporter/src/github_client"
)

var (
  botToken = os.Getenv("BOTTOKEN")
  users = make(Users)
)

type Users map[int64]*UserState

type UserState struct {
  GitHubURLs []string
}

type Bot struct {
  tgbotapi.BotAPI
  githubClient *github_client.GitHubClient
}

func initBot() (*Bot, error) {
  bot, err := tgbotapi.NewBotAPI(botToken)
  if err != nil {
    return nil, err
  }

  githubClient := github_client.newGitHubClient()


  return &Bot{&bot, &githubClient}, nil
}

func (bot *Bot) startPolling() {
  u := tgbotapi.NewUpdate(0)
  u.Timeout = 60

  updates := bot.GetUpdatesChan(u)

  for update := range updates {
    if update.Message != nil {
      bot.handleMessage()
    } else if update.CallbackQuery != nil {
      bot.handleCallbackQuery()
    }

    bot.Send(msg)
  }
}

func (bot *Bot) handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, githubClient *GitHubClient) {
  chatID := message.Chat.ID

  if userState, exists := userStates[chatID]; exists {
    return
  }

  msg := tgbotapi.NewMessage(chatID, "")
  var githubURLs []string

  for _, repoURL := range strings.Split(message.Text, "\n") {
    append(githubURLs, repoURL) 
  }

  userState[chatID] = &UserState{GitHubURLs: githubURLs}

  keyboard := tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
      tgbotapi.NewInlineKeyboardButtonData("1 day", "1_day"),
      tgbotapi.NewInlineKeyboardButtonData("1 week", "1_week"),
      tgbotapi.NewInlineKeyboardButtonData("1 month", "1_month"),
    ),
  )

  msg := tgbotapi.NewMessage(chatID, "Choose time interval:")
  msg.ReplyMarkup = keyboard

  bot.Send(msg)
}

func (bot *Bot) handleCallbackQuery() {
  chatID := callbackQuery.Message.Chat.ID

  callback := tgbotapi.NewCallback(callbackQuery.ID, "You selected: "+callbackQuery.Data)

  if _, err := bot.Request(callback); err != nil {
    log.Println("Error responding to callback query:", err)
  }

  if userState, exists := userStates[chatID]; !exists || len(userState.GitHubURLs) == 0 {
    msg := tgbotapi.NewMessage(chatID, "No GitHub links found, please provide ones.") 
    bot.Send(msg)
    return
  }

  msg := tgbotapi.NewMessage(chatID, "")

  for repoURL := range userState.GitHubURLs {
    repoInfo, err := githubClient.repoInfo(repoURL)

    switch err.(AppError) {
    case InvalidGitHubURL:
      msg.Text += "Invalid github URL: " + repoURL
    case RepoFetchingError:
      msg.Text += "Error fetching repository: " + repoURL
    default:
      msg.Text += repoInfo
    }

    msg.Text += "\n\n"
  }

  bot.Sent(msg)
  delete(userState[chatID])
}

func isValidRepoURL(repoURL string) bool {
  return strings.HasPrefix(repoURL, "https://github.com/") && len(strings.Split(repoURL, "/")) == 5
}
