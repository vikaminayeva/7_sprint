package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {

	city := "moscow"

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: len(cafeList[city])},
	}

	for _, c := range requests {
		response := httptest.NewRecorder()

		target := fmt.Sprintf("/cafe?city=%s&count=%d", city, c.count)
		req := httptest.NewRequest("GET", target, nil)

		handler := http.HandlerFunc(mainHandle)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := response.Body.String()

		list := strings.Split(body, ",")
		if c.count == 0 {
			assert.Equal(t, 0, len(body)) // если ждем 0, то и тело должно быть пустым
		} else {
			assert.Equal(t, c.want, len(list))
		}
	}
}

func TestCafeSearch(t *testing.T) {
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}
	for _, c := range requests {
		response := httptest.NewRecorder()

		target := fmt.Sprintf("/cafe?city=%s&search=%s", city, c.search)
		req := httptest.NewRequest("GET", target, nil)

		handler := http.HandlerFunc(mainHandle)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := response.Body.String()

		var list []string
		if body != "" {
			list = strings.Split(body, ",")
		}
		assert.Len(t, list, c.wantCount)

		for _, cafe := range list {
			cafeLower := strings.ToLower(cafe)
			searchLower := strings.ToLower(c.search)
			assert.True(t, strings.Contains(cafeLower, searchLower), "Кафе %s должно содержать %s", cafe, c.search)
		}
	}
}
