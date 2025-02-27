package main

import (
  "os"
  "strings"
  "fmt"
  "log"
  "context"
  "github.com/google/go-github/v69/github"
  tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
  "golang.org/x/oauth2"
)

var (
  gitToken = os.Getenv("TOKEN")
  botToken = os.Getenv("BOTTOKEN")
)

func main() {
  ctx := context.Background()
  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: gitToken},
  )
  tc := oauth2.NewClient(ctx, ts)
  githubClient := github.NewClient(tc)

  bot, err := tgbotapi.NewBotAPI(botToken)
  if err != nil {
    log.Panic(err)
  }

  bot.Debug = true

  log.Printf("Authorized on account %s", bot.Self.UserName)

  u := tgbotapi.NewUpdate(0)
  u.Timeout = 60
  
  updates := bot.GetUpdatesChan(u)

  for update := range updates {
    if strings.Contains(update.Message.Text, "github.com") {
      repoURL := update.Message.Text
      repoInfo := getGitHubRepoInfo(githubClient, ctx, repoURL) 

      msg := tgbotapi.NewMessage(update.Message.Chat.ID, repoInfo)
      bot.Send(msg)
    } else {
      msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please send valid github repo link")
      bot.Send(msg)
    }
  }
}

func getGitHubRepoInfo(client *github.Client, ctx context.Context, repoURL string) string {
  parts := strings.Split(repoURL, "/")
  if len(parts) < 5 {
    return "Invalid GitHub repository URL"
  }

  owner := parts[3]
  repo := parts[4]

  repository, _, err := client.Repositories.Get(ctx, owner, repo)
  if err != nil {
    return fmt.Sprintf("Error fetching repository info: %v", err)
  }

  info := fmt.Sprintf("Repository: %s\n", *repository.FullName)
	info += fmt.Sprintf("Description: %s\n", *repository.Description)
	info += fmt.Sprintf("Stars: %d\n", *repository.StargazersCount)
	info += fmt.Sprintf("Forks: %d\n", *repository.ForksCount)
	info += fmt.Sprintf("Watchers: %d\n", *repository.WatchersCount)
	info += fmt.Sprintf("Language: %s\n", *repository.Language)
	info += fmt.Sprintf("URL: %s\n", *repository.HTMLURL)

  return info
}
