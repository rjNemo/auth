package server

// PageData contains fields shared by the templates for now.
type PageData struct {
	Email string
	Error string
}

func newIndexData(email, errMsg string) PageData {
	return PageData{Email: email, Error: errMsg}
}

func newUnauthorizedData(errMsg string) PageData {
	return PageData{Error: errMsg}
}
