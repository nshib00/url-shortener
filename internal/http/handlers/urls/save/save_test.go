package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-url-shortener/internal/http/handlers/urls/save"
	"go-url-shortener/internal/http/handlers/urls/save/mocks"
	"go-url-shortener/internal/http/testutils"
	"go-url-shortener/internal/utils/logger/handlers/mocklogger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name          string
		alias         string
		url           string
		userID        int
		respError     string
		mockError     error
		expectedCalls bool
		expectedCode  int
	}{
		{
			name:          "Success",
			alias:         "test_alias",
			url:           "https://google.com",
			userID:        12345,
			expectedCalls: true,
			expectedCode:  http.StatusCreated,
		},
		{
			name:          "Empty alias",
			alias:         "",
			url:           "https://google.com",
			userID:        12345,
			expectedCalls: true,
			expectedCode:  http.StatusCreated,
		},
		{
			name:          "Empty URL",
			url:           "",
			alias:         "some_alias",
			userID:        12345,
			respError:     "field URL is required",
			expectedCalls: false,
			expectedCode:  http.StatusOK, // ошибка валидации возвращает 200
		},
		{
			name:          "Invalid URL",
			url:           "some invalid URL",
			alias:         "some_alias",
			userID:        12345,
			respError:     "field URL is not a valid URL",
			expectedCalls: false,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "SaveURL Error",
			alias:         "test_alias",
			url:           "https://google.com",
			userID:        12345,
			respError:     "failed to add url",
			mockError:     errors.New("unexpected error"),
			expectedCalls: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:         "Unauthorized - missing auth header",
			alias:        "test_alias",
			url:          "https://google.com",
			userID:       0,
			respError:    "missing auth header",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Unauthorized - invalid token",
			alias:        "test_alias",
			url:          "https://google.com",
			userID:       -1,
			respError:    "invalid token",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			urlSaverMock := mocks.NewURLSaver(tt)

			if tc.expectedCalls && tc.userID > 0 {
				if tc.alias == "" {
					urlSaverMock.On("SaveURL", tc.userID, tc.url, mock.AnythingOfType("string")).
						Return(int64(1), tc.mockError).
						Once()
				} else {
					urlSaverMock.On("SaveURL", tc.userID, tc.url, tc.alias).
						Return(int64(1), tc.mockError).
						Once()
				}
			}

			handler := save.New(mocklogger.NewMockLogger(), urlSaverMock)

			r := chi.NewRouter()

			authMiddleware := testutils.CreateAuthMiddleware()
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)
				r.Post("/save", handler.Handle)
			})

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			req, err := http.NewRequest(
				http.MethodPost,
				"/save",
				bytes.NewReader([]byte(input)),
			)
			require.NoError(tt, err)

			if tc.userID > 0 {
				token := testutils.GenerateTestToken(tt, tc.userID)
				req.Header.Set("Authorization", "Bearer "+token)
			} else if tc.userID == -1 {
				req.Header.Set("Authorization", "Bearer invalid_token")
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(tt, tc.expectedCode, rec.Code)

			if tc.respError == "missing auth header" || tc.respError == "invalid token" {
				var errorResp map[string]string
				require.NoError(tt, json.Unmarshal(rec.Body.Bytes(), &errorResp))
				require.Equal(tt, tc.respError, errorResp["error"])
			} else {
				var resp save.Response
				require.NoError(tt, json.Unmarshal(rec.Body.Bytes(), &resp))
				require.Equal(tt, tc.respError, resp.Error)

				if tc.respError == "" && tc.expectedCalls {
					if tc.alias != "" {
						require.Equal(tt, tc.alias, resp.Alias)
					} else {
						require.NotEmpty(tt, resp.Alias)
						require.Len(tt, resp.Alias, 8) // aliasLength = 8
					}
				}
			}
		})
	}
}
