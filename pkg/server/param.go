package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func extractString(r *http.Request, key string) (value string, err error) {
	if value = r.FormValue(key); value == "" {
		err = paramNotFound(key)
		return
	}
	return
}

func extractBool(r *http.Request, key string, allowNotFound bool) (value bool, err error) {
	var strVal string
	if strVal, err = extractString(r, key); err != nil {
		if allowNotFound {
			return false, nil
		}
		return
	}
	if value, err = strconv.ParseBool(strVal); err != nil {
		err = paramInvalid(key)
		return
	}
	return
}

func extractInt(r *http.Request, key string) (value int, err error) {
	var strVal string
	if strVal, err = extractString(r, key); err != nil {
		return
	}
	if value, err = strconv.Atoi(strVal); err != nil {
		err = paramInvalid(key)
		return
	}
	return
}

func extractInt64(r *http.Request, key string) (value int64, err error) {
	var strVal string
	if strVal, err = extractString(r, key); err != nil {
		return
	}
	if value, err = strconv.ParseInt(strVal, 10, 64); err != nil {
		err = paramInvalid(key)
	}
	return
}

func extractUint64(r *http.Request, key string) (value uint64, err error) {
	var strVal string
	if strVal, err = extractString(r, key); err != nil {
		return
	}
	if value, err = strconv.ParseUint(strVal, 10, 64); err != nil {
		err = paramInvalid(key)
		return
	}
	return
}

func paramNotFound(name string) (err error) {
	return fmt.Errorf("parameter %v not found", name)
}

func paramInvalid(name string) (err error) {
	return fmt.Errorf("parameter %v invalid", name)
}
