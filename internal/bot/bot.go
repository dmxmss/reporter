package bot

import (
  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
  "strings"
  "os"
  "reporter/internal/github_client"
  "reporter/internal/app_error"
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
  *tgbotapi.BotAPI
  githubClient *github_client.GitHubClient
}

func InitBot() (*Bot, error) {
  bot, err := tgbotapi.NewBotAPI(botToken)
  if err != nil {
    return nil, err
  }

  githubClient := github_client.NewGitHubClient()

  return &Bot{bot, githubClient}, nil
}

func (bot *Bot) StartPolling() {
  u := tgbotapi.NewUpdate(0)
  u.Timeout = 60

  updates := bot.GetUpdatesChan(u)

  for update := range updates {
    if update.Message != nil {
      if update.Message.IsCommand() {
        bot.handleCommand(update.Message)
      } else {
        bot.handleMessage(update.Message)
      }
    } else if update.CallbackQuery != nil {
      bot.handleCallbackQuery(update.CallbackQuery)
    }
  }
}

func (bot *Bot) handleCommand(message *tgbotapi.Message) {
  command := message.Command()
  chatID := message.Chat.ID

  msg := tgbotapi.NewMessage(chatID, "")
  switch command {
  case "start":
    msg.Text = "Hello, I am reporter bot. Give me github repositories links and choose the period and I show you commits for this period of time." 
  case "help":
    msg.Text = `
<repo URL 1>
<repo URL 2>
...
    `
  default:
    msg.Text = "I do not know this command."
  }

  bot.Send(msg)
}

func (bot *Bot) handleMessage(message *tgbotapi.Message) {
  chatID := message.Chat.ID

  if _, exists := users[chatID]; exists {
    return
  }

  var urls []string

  for _, repoURL := range strings.Split(message.Text, "\n") {
    if !isValidRepoURL(repoURL) {
      msg := tgbotapi.NewMessage(chatID, "Invalid github repository link: " + repoURL)
      bot.Send(msg)
      continue
    }

    urls = append(urls, repoURL) 
  }

  users[chatID] = &UserState{GitHubURLs: urls}

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

func (bot *Bot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
  chatID := callbackQuery.Message.Chat.ID

  var days int
  switch callbackQuery.Data {
  case "1_day": 
    days = 1
  case "1_week":
    days = 7
  case "1_month":
    days = 31
  default:
    days = 1
  }

  userState, exists := users[chatID]

  if !exists || len(userState.GitHubURLs) == 0 {
    msg := tgbotapi.NewMessage(chatID, "No GitHub links found, please provide ones.") 
    bot.Send(msg)
    return
  }

  msg := tgbotapi.NewMessage(chatID, "")

  for _, repoURL := range userState.GitHubURLs {
    commitsInfo, err := bot.githubClient.CommitInfoSince(repoURL, days)

    if err != nil {
      switch err {
      case app_error.InvalidGitHubURL:
        msg.Text += "Invalid github URL: " + repoURL
      case app_error.RepoFetchingError:
        msg.Text += "Error fetching repository: " + repoURL
      case app_error.CommitFetchingError:
        msg.Text += "Error fetching commits" 
      }

      bot.Send(msg)
      continue
    }

    msg.Text = commitsInfo

    bot.Send(msg)
  }

  delete(users, chatID)
}

func isValidRepoURL(repoURL string) bool {
  return strings.HasPrefix(repoURL, "https://github.com/") && len(strings.Split(repoURL, "/")) == 5
}
