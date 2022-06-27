package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "election/db/mock"
	db "election/db/sqlc"
	"election/util"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	fmt.Printf("Expected: %v and Actual: %v", e.arg, arg)

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func eqCreateUserParams(password string, arg db.CreateUserParams) gomock.Matcher {
	return eqCreateUserParamsMatcher{
		arg:      arg,
		password: password,
	}
}

func TestCreateUserAPI(t *testing.T) {

	user, password := CreateRandomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
				"full_name":   user.FullName,
				"email":       user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.CreateUserParams{
					NationalID: user.NationalID,
					FullName:   user.FullName,
					Email:      user.Email,
					Permission: user.Permission,
					HasVoted:   false,
				}

				store.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(password, arg)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
				"full_name":   user.FullName,
				"email":       user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateNationalID",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
				"full_name":   user.FullName,
				"email":       user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidNationalID",
			body: gin.H{
				"national_id": "invalid#1",
				"password":    password,
				"full_name":   user.FullName,
				"email":       user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
				"full_name":   user.FullName,
				"email":       "invalid-email",
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    "123",
				"full_name":   user.FullName,
				"email":       user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			tc.buildStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/users")
			values, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(values))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := CreateRandomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IncorrectPassword",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    "incorrect",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"national_id": user.NationalID,
				"password":    password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidNationalID",
			body: gin.H{
				"national_id": "invalid-national_id#1",
				"password":    password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, users db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, users.NationalID, gotUser.NationalID)
	require.Equal(t, users.FullName, gotUser.FullName)
	require.Equal(t, users.Email, gotUser.Email)
	require.Equal(t, users.Permission, gotUser.Permission)
	require.Equal(t, users.HasVoted, gotUser.HasVoted)
	require.Empty(t, gotUser.HashedPassword)
}

func CreateRandomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	permission := make([]string, 1)
	permission[0] = util.Vote

	user = db.User{
		NationalID:     "1234567890123",
		HashedPassword: hashedPassword,
		FullName:       util.RandomName(),
		Email:          util.RandomEmail(),
		Permission:     permission,
		HasVoted:       false,
	}
	return
}
