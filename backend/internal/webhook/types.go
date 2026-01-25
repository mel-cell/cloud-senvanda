package webhook

// GiteaPushPayload represents the JSON payload sent by Gitea on push events.
// We only map fields that are essential for our deployment logic.
type GiteaPushPayload struct {
	Ref        string     `json:"ref"`    // e.g., "refs/heads/main"
	Before     string     `json:"before"` // Commit hash before push
	After      string     `json:"after"`  // Commit hash after push (The Target)
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Sender     Sender     `json:"sender"`
}

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"` // e.g., "user/repo"
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
	SSHURL   string `json:"ssh_url"`
}

type Pusher struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Sender struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
