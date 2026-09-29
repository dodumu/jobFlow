package utils

import (
	"net/http"
	"strconv"
)

func GetPathID(r *http.Request, name string) (int, error) {
	value := r.PathValue(name)

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, strconv.ErrSyntax
	}

	return id, nil
}
