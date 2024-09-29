package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"simple_bank/constants"
	"simple_bank/middleware"
	"simple_bank/pkg/token"
	"testing"
	"time"
)

func addMiddleware(
	t *testing.T,
	request *http.Request,
	authWebTokenType string,
	tokenMaker token.Maker,
	username string,
	duration time.Duration,
) {
	tokenString, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	authorizationHeader := fmt.Sprintf("%s %s", authWebTokenType, tokenString)
	request.Header.Set(constants.AuthorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, token token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

			},
		},
		{
			name: "空授权类型",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, "", tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "其它授权类型",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, "OAuth", tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "令牌过期",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addMiddleware(t, req, constants.AuthorizationHeaderType, tokenMaker, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				middleware.AuthWebTokenMiddleware(server.tokenMake),
				func(ctx *gin.Context,
				) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMake)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}
