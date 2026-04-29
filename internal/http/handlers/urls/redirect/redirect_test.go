package redirect_test

import (
	"go-url-shortener/internal/http/handlers/urls/redirect"
	"go-url-shortener/internal/http/handlers/urls/redirect/mocks"
	"go-url-shortener/internal/http/testutils"
	"go-url-shortener/internal/storage"
	"go-url-shortener/internal/utils/logger/handlers/mocklogger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name          string
		alias         string
		url           string
		userID        int
		respError     string
		mockError     error
		expectedCalls bool
		needAuth      bool
	}{
		{
			name:          "Success - with auth",
			alias:         "test",
			url:           "https://www.google.com/",
			userID:        12345,
			expectedCalls: true,
			needAuth:      true,
		},
		{
			name:          "Success - without auth (public route)",
			alias:         "test-public",
			url:           "https://www.google.com/",
			userID:        0,
			expectedCalls: true,
			needAuth:      false,
		},
		{
			name:          "Empty alias",
			alias:         "",
			url:           "https://www.google.com/",
			userID:        12345,
			expectedCalls: false,
			needAuth:      true,
		},
		{
			name:          "Empty URL",
			alias:         "empty-url",
			url:           "",
			userID:        12345,
			respError:     "invalid request: alias is empty",
			mockError:     storage.ErrURLNotFound,
			expectedCalls: true,
			needAuth:      true,
		},
		{
			name:          "Alias with dot",
			alias:         "my.alias",
			url:           "https://www.google.com/",
			userID:        12345,
			expectedCalls: true,
			needAuth:      true,
		},
		{
			name:          "Random string in URL passed",
			alias:         "random-str",
			url:           "some-string-but-not-url",
			userID:        12345,
			respError:     "invalid url",
			expectedCalls: true,
			needAuth:      true,
		},
		{
			name:          "Short alias",
			alias:         "a",
			url:           "https://www.google.com/",
			userID:        12345,
			expectedCalls: true,
			needAuth:      true,
		},
		{
			name:      "Unauthorized - missing auth header",
			alias:     "protected",
			userID:    0,
			respError: "missing auth header",
			needAuth:  true,
		},
		{
			name:      "Unauthorized - invalid token",
			alias:     "protected",
			userID:    -1, // специальное значение для невалидного токена
			respError: "invalid token",
			needAuth:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			if tc.alias == "" {
				handler := redirect.New(mocklogger.NewMockLogger(), mocks.NewURLGetter(tt))
				r := chi.NewRouter()
				r.Get("/{alias}", handler.Handle)

				ts := httptest.NewServer(r)
				defer ts.Close()

				res, err := http.Get(ts.URL + "/")
				require.NoError(tt, err)
				assert.Equal(tt, http.StatusNotFound, res.StatusCode)
				return
			}

			urlGetterMock := mocks.NewURLGetter(tt)

			if tc.expectedCalls && tc.userID >= 0 {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			handler := redirect.New(mocklogger.NewMockLogger(), urlGetterMock)
			r := chi.NewRouter()

			if tc.needAuth {
				authMiddleware := testutils.CreateAuthMiddleware()
				r.Group(func(r chi.Router) {
					r.Use(authMiddleware)
					r.Get("/{alias}", handler.Handle)
				})
			} else {
				r.Get("/{alias}", handler.Handle)
			}

			ts := httptest.NewServer(r)
			defer ts.Close()

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			req, err := http.NewRequest("GET", ts.URL+"/"+tc.alias, nil)
			require.NoError(tt, err)

			if tc.userID > 0 {
				token := testutils.GenerateTestToken(tt, tc.userID)
				req.Header.Set("Authorization", "Bearer "+token)
			} else if tc.userID == -1 {
				req.Header.Set("Authorization", "Bearer invalid_token")
			}

			res, err := client.Do(req)
			require.NoError(tt, err)
			defer res.Body.Close()

			if tc.respError == "missing auth header" || tc.respError == "invalid token" {
				assert.Equal(tt, http.StatusUnauthorized, res.StatusCode)
				return
			}

			if tc.alias == "" {
				assert.Equal(tt, http.StatusNotFound, res.StatusCode)
				return
			}

			if tc.respError != "" && tc.respError != "missing auth header" && tc.respError != "invalid token" {
				assert.Equal(tt, http.StatusOK, res.StatusCode)
				return
			}
			assert.Equal(tt, http.StatusFound, res.StatusCode)
			assert.Equal(tt, tc.url, res.Header.Get("Location"))
		})
	}
}
