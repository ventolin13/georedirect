package messages

type RedirectRequest struct {
	ClientIP    string
	UserAgent   string
	Fingerprint string
	Path        string
	Query       string
}

type RedirectResponse struct {
	Location string
	Blocked  bool
}
