// Copyright 2014 Simon Zimmermann. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httputil

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var (
	invalidValue = reflect.Value{}
)

func ConvertTime(value string) reflect.Value {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return reflect.ValueOf(t)
	}
	return invalidValue
}

func OffsetCount(r *http.Request) (int, int) {
	var count, offset uint64
	var err error

	if count, err = strconv.ParseUint(r.FormValue("count"), 10, 0); err != nil {
		count = 10
	}

	if offset, err = strconv.ParseUint(r.FormValue("offset"), 10, 0); err != nil {
		offset = 0
	}

	return int(offset), int(count)
}

func FormValue(r *http.Request, key string) (string, error) {
	vs, ok := r.Form[key]

	if !ok || len(vs) == 0 {
		return "", fmt.Errorf("key does not exist %s", key)
	}

	return vs[0], nil
}

func FormValues(r *http.Request, key string) ([]string, error) {
	vs, ok := r.Form[key]

	if !ok || len(vs) == 0 {
		return nil, fmt.Errorf("key does not exist %s", key)
	}

	return vs, nil
}

func ParseUint(r *http.Request, key string) (uint, error) {
	v, err := strconv.ParseUint(r.FormValue(key), 10, 0)
	return uint(v), err
}

func ParseInt(r *http.Request, key string) (int, error) {
	v, err := strconv.ParseInt(r.FormValue(key), 10, 0)
	return int(v), err
}

func ParseIntArray(r *http.Request, key string) ([]int, error) {
	v, err := FormValues(r, key)
	if err != nil {
		return nil, err
	}
	vint := make([]int, len(v))

	for i := range v {
		if vi, err := strconv.ParseInt(v[i], 10, 0); err != nil {
			return nil, err
		} else {
			vint[i] = int(vi)
		}
	}

	return vint, err
}

func ParseBool(r *http.Request, key string) (bool, error) {
	return strconv.ParseBool(r.FormValue(key))
}

func ParseString(r *http.Request, key string) (string, error) {
	return FormValue(r, key)
}

func SetUint(r *http.Request, m map[string]interface{}, key string) error {
	v, err := ParseUint(r, key)
	
	if err != nil {
		return err
	}

	m[key] = v
	return nil
}

func SetInt(r *http.Request, m map[string]interface{}, key string) error {
	v, err := ParseInt(r, key)
	
	if err != nil {
		return err
	}

	m[key] = v
	return nil
}

func SetIntArray(r *http.Request, m map[string]interface{}, key string) error {
	v, err := ParseIntArray(r, key)
	
	if err != nil {
		return err
	}

	m[key] = v
	return nil
}

func SetString(r *http.Request, m map[string]interface{}, key string) error {
	v, err := FormValue(r, key)

	if err != nil {
		return err
	}

	m[key] = v
	return nil
}

func SetBool(r *http.Request, m map[string]interface{}, key string) error {
	v, err := ParseBool(r, key)

	if err != nil {
		return err
	}

	m[key] = v
	return nil
}
