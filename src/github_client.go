package github_client
import (
  "context"
  "github.com/google/go-github/v69/github"
  "golang.org/x/oauth2"
)

var (
  gitToken = os.Getenv("TOKEN")
)

type GitHubClient struct {
  ctx *context.Context
  client *github.Client
}

func newGitHubClient() *GitHubClient {
  ctx := context.Background()

  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: gitToken},
  )
  tc := oauth2.NewClient(ctx, ts)
  githubClient := github.NewClient(tc)

  return &GitHubClient{ctx: &ctx, client: &githubClient}
}

func (githubClient *GitHubClient) fetchRepo(owner string, repo string) (*github.Repository, error) {
  repository, _, err := githubClient.client.Repositories.Get(githubClient.ctx, owner, repo)
  if err != nil {
    return nil, RepoFetchingError
  }

  return repository, nil
}

func (githubClient *GitHubClient) repoInfo(repoURL string) (string, error) {
  parts := strings.Split(repoURL, "/")
  if len(parts) < 5 {
    return "", InvalidGitHubURL
  }

  owner := parts[3]
  repo := parts[4]

  repository, err := githubClient.fetchRepo(owner, repo)
  if err != nil {
    return "", RepoFetchingError
  }

  info := formatGitHubRepoInfo(repository)

  return info, nil
}

func formatGitHubRepoInfo(repository *github.Repository) string {
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

  return info
}

