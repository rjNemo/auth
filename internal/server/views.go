package server

// PageData contains fields shared by the templates for now.
type PageData struct {
	Email     string
	Error     string
	CSRFToken string
}

func newIndexData(email, errMsg, token string) PageData {
	return PageData{Email: email, Error: errMsg, CSRFToken: token}
}

func newUnauthorizedData(errMsg, token string) PageData {
	return PageData{Error: errMsg, CSRFToken: token}
}
