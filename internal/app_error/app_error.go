package app_error

type AppError int

const (
  InvalidGitHubURL AppError = iota
  RepoFetchingError
  CommitFetchingError
)

func (err AppError) Error() string {
  var message string

  switch err {
  case InvalidGitHubURL:
    message = "Invalid github URL"
  case RepoFetchingError:
    message = "Repository fetching error"
  case CommitFetchingError:
    message = "Commit fetching error"
  }

  return message
}
