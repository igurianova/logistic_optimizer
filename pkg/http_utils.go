package pkg

import "net/http"

func WriteHttpStatusResponse(w http.ResponseWriter, status int) error {
	return WriteHttpMessageResponse(w, http.StatusText(status), status)
}

func WriteHttpMessageResponse(w http.ResponseWriter, msg string, status int) error {
	return WriteHttpByteResponse(w, []byte(msg), status)
}

func WriteHttpByteResponse(w http.ResponseWriter, bytes []byte, status int) error {
	w.WriteHeader(status)
	_, err := w.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
