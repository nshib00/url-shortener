package redirect_test

import (
	"go-url-shortener/internal/http/handlers/urls/redirect"
	"go-url-shortener/internal/http/handlers/urls/redirect/mocks"
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
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test",
			url:   "https://www.google.com/",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://www.google.com/",
		},
		{
			name:      "Empty URL",
			alias:     "empty-url",
			url:       "",
			respError: "invalid request: alias is empty",
			mockError: storage.ErrURLNotFound,
		},
		{
			name:  "Alias with dot",
			alias: "my.alias",
			url:   "https://www.google.com/",
		},
		{
			name:      "Random string in URL passed",
			alias:     "random-str",
			url:       "some-string-but-not-url",
			respError: "invalid url",
		},
		{
			name:  "Short alias",
			alias: "a",
			url:   "https://www.google.com/",
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
			if tc.alias != "" {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			handler := redirect.New(mocklogger.NewMockLogger(), urlGetterMock)
			r := chi.NewRouter()
			r.Get("/{alias}", handler.Handle)

			ts := httptest.NewServer(r)
			defer ts.Close()

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			res, err := client.Get(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			if tc.alias == "" {
				res, err := client.Get(ts.URL + "/")
				require.NoError(tt, err)
				assert.Equal(tt, http.StatusNotFound, res.StatusCode)
				return
			}
			if tc.respError != "" {
				assert.Equal(tt, http.StatusOK, res.StatusCode)
				return
			}

			assert.Equal(t, http.StatusFound, res.StatusCode) // 302
			assert.Equal(t, tc.url, res.Header.Get("Location"))
		})
	}

}
