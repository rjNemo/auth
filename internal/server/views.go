package server

// PageData contains fields shared by the templates for now.
type PageData struct {
	Email        string
	Error        string
	Info         string
	CSRFToken    string
	CreatedAt    string
	CreatedAtISO string
}

func newIndexData(email, errMsg, token string) PageData {
	return PageData{Email: email, Error: errMsg, CSRFToken: token}
}

func newUnauthorizedData(errMsg, token string) PageData {
	return PageData{Error: errMsg, CSRFToken: token}
}

func newDashboardData(email, token, createdAt, createdAtISO string) PageData {
	return PageData{Email: email, CSRFToken: token, CreatedAt: createdAt, CreatedAtISO: createdAtISO}
}

func newSignupData(email, errMsg, token string) PageData {
	return PageData{Email: email, Error: errMsg, CSRFToken: token}
}
