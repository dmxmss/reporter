package github_client
import (
  "context"
  "strings"
  "os"
  "time"
  "fmt"
  "github.com/google/go-github/v69/github"
  "golang.org/x/oauth2"
  "reporter/internal/app_error"
)

var (
  gitToken = os.Getenv("TOKEN")
)

type GitHubClient struct {
  *github.Client
  Context *context.Context
}

func NewGitHubClient() *GitHubClient {
  ctx := context.Background()

  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: gitToken},
  )
  tc := oauth2.NewClient(ctx, ts)
  githubClient := github.NewClient(tc)

  return &GitHubClient{Context: &ctx, Client: githubClient}
}

func (githubClient *GitHubClient) fetchRepo(repoURL string) (*github.Repository, error) {
  owner, repo, err := extractCredentials(repoURL)
  if err != nil {
    return nil, err
  }

  repository, _, err := githubClient.Repositories.Get(*githubClient.Context, owner, repo)
  if err != nil {
    return nil, app_error.RepoFetchingError
  }

  return repository, nil
}

func (githubClient *GitHubClient) fetchCommitsSince(repoURL string, days int) ([]*github.RepositoryCommit, error) {
  since := time.Now().AddDate(0, 0, -days)

  opt := &github.CommitsListOptions {
    Since: since,
    ListOptions: github.ListOptions {
      PerPage: 100,
    },
  }

  var commits[]*github.RepositoryCommit
  owner, repo, err := extractCredentials(repoURL)
  if err != nil {
    return nil, err 
  }

  for {
    pageCommits, resp, err := githubClient.Repositories.ListCommits(*githubClient.Context, owner, repo, opt)
    if err != nil {
      return nil, app_error.CommitFetchingError
    }

    commits = append(commits, pageCommits...)

    if resp.NextPage == 0 {
      break
    }
  }

  return commits, nil
}

func (githubClient *GitHubClient) CommitInfoSince(repoURL string, days int) (string, error) {
  commits, err := githubClient.fetchCommitsSince(repoURL, days)
  if err != nil {
    return "", app_error.CommitFetchingError
  }

  if len(commits) == 0 {
    return "There are no commits for this period", nil
  }

  info := repoURL + ":\n"

  for _, commit := range commits {
    info += formatCommitInfo(commit.GetCommit().GetMessage(), commit.GetCommit().GetAuthor().GetDate().UTC().Format(time.DateTime), commit.GetURL()) + "\n\n"
  }

  return info, nil
}

func formatCommitInfo(message, date, url string) string {
  return fmt.Sprintf("%s, %s \n(%s)", message, date, url)
}

func (githubClient *GitHubClient) RepoInfo(repoURL string) (string, error) {
  repository, err := githubClient.fetchRepo(repoURL)
  if err != nil {
    return "", app_error.RepoFetchingError
  }

  info := fmt.Sprintf("Repository: %s\n", *repository.FullName)
  if(repository.Description != nil) {
	  info += fmt.Sprintf("Description: %s\n", *repository.Description) 
  } else {
    info += "Description: <empty>\n"
  }
	info += fmt.Sprintf("Stars: %d\n", *repository.StargazersCount)
	info += fmt.Sprintf("Forks: %d\n", *repository.ForksCount)
	info += fmt.Sprintf("Watchers: %d\n", *repository.WatchersCount)
	info += fmt.Sprintf("Language: %s\n", *repository.Language)

  return info, nil
}

func extractCredentials(repoURL string) (string, string, error) {
  parts := strings.Split(repoURL, "/")
  if len(parts) < 5 {
    return "", "", app_error.InvalidGitHubURL
  }

  owner := parts[3]
  repo := parts[4]

  return owner, repo, nil
}
