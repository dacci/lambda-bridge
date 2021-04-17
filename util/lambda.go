package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func InvokeLambda(event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/2015-03-31/functions/function/invocations", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
