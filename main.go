package main

import (
  "os"
  "strings"
  // "errors"
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
  githubClient := getGitHubClient(ctx)

  bot, err := tgbotapi.NewBotAPI(botToken)
  if err != nil {
    log.Fatal(err)
  }

  bot.Debug = true

  log.Printf("Authorized on account %s", bot.Self.UserName)

  err = startPolling(bot, githubClient, ctx)
  if err != nil {
    log.Fatal(err)
  }
}

func getGitHubClient(ctx context.Context) *github.Client {
  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: gitToken},
  )
  tc := oauth2.NewClient(ctx, ts)
  githubClient := github.NewClient(tc)

  return githubClient
}

func startPolling(bot *tgbotapi.BotAPI, githubClient *github.Client, ctx context.Context) error {
  u := tgbotapi.NewUpdate(0)
  u.Timeout = 60

  updates := bot.GetUpdatesChan(u)

  for update := range updates {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
    for _, repoURL := range strings.Split(update.Message.Text, "\n") {
      if isValidRepoURL(repoURL) {
        repoInfo, err := getGitHubRepoInfo(githubClient, ctx, repoURL) 
        if err != nil {
          msg.Text += repoURL + " not found"
          break
        }

        msg.Text += repoInfo
      } else {
        msg.Text += "Invalid github repo link: " + repoURL
        break
      }

      msg.Text += "\n\n"
    }

    bot.Send(msg)
  }

  return nil
}

func isValidRepoURL(repoURL string) bool {
  return strings.HasPrefix(repoURL, "https://github.com/") && len(strings.Split(repoURL, "/")) == 5
}

func getGitHubRepoInfo(client *github.Client, ctx context.Context, repoURL string) (string, error) {
  parts := strings.Split(repoURL, "/")

  owner := parts[3]
  repo := parts[4]

  repository, _, err := client.Repositories.Get(ctx, owner, repo)
  if err != nil {
    return "", err
  }

  info := fmt.Sprintf("Repository: %s\n", *repository.FullName)
  if(repository.Description != nil) {
	  info += fmt.Sprintf("Description: %s\n", *repository.Description) 
  }
	info += fmt.Sprintf("Stars: %d\n", *repository.StargazersCount)
	info += fmt.Sprintf("Forks: %d\n", *repository.ForksCount)
	info += fmt.Sprintf("Watchers: %d\n", *repository.WatchersCount)
	info += fmt.Sprintf("Language: %s\n", *repository.Language)

  return info, nil
}
