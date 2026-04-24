package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-url-shortener/internal/http/handlers/save/mocks"
	"go-url-shortener/internal/utils/logger/handlers/mocklogger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:      "Empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is required",
		},
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			urlSaverMock := mocks.NewURLSaver(tt)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := New(mocklogger.NewMockLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			req, err := http.NewRequest(
				http.MethodPost,
				"/save",
				bytes.NewReader([]byte(input)),
			)
			require.NoError(tt, err)

			rec := httptest.NewRecorder()
			handler.Handle(rec, req)

			require.Equal(tt, rec.Code, http.StatusOK)
			body := rec.Body.String()

			var resp Response
			require.NoError(tt, json.Unmarshal([]byte(body), &resp))
			require.Equal(tt, tc.respError, resp.Error)
		})
	}
}
