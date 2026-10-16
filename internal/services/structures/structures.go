package structures

type Click struct {
	ClientIP    string
	UserAgent   string
	Fingerprint string
	Path        string
	Query       string
}

type Decision struct {
	Location string
	Blocked  bool
}
