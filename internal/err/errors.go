package err

type APIError interface {
	// APIError возвращает HTTP статус и ошибку.
	APIError() (int, string)
}

type HTTPError struct {
	Msg    string
	Status int
}

func (e HTTPError) Error() string {
	return e.Msg
}

func (e HTTPError) APIError() (int, string) {
	return e.Status, e.Msg
}
