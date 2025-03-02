package AppError

type AppError int

const (
  InvalidGitHubURL AppError = iota
  RepoFetchingError
)

func (err AppError) Error() string {
  var message string

  switch err {
  case InvalidGitHubURL:
    message = "Invalid github URL"
  case RepoFetchingError:
    message = "Repository fetching error"
  }

  return message
}
