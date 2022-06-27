package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "election/db/mock"

	db "election/db/sqlc"

	"election/token"
	"election/util"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateCandidateAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	candidate := RandomCandidate()
	rspCandidate := NewCandidateResponse(candidate)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":      candidate.Name,
				"dob":       candidate.Dob,
				"bioLink":   candidate.BioLink,
				"imageLink": candidate.ImageUrl,
				"policy":    candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.CreateCandidateParams{
					Name:      candidate.Name,
					Dob:       candidate.Dob,
					BioLink:   candidate.BioLink,
					ImageUrl:  candidate.ImageUrl,
					Policy:    candidate.Policy,
					VoteCount: 0,
				}

				store.EXPECT().
					CreateCandidate(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(candidate, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateResponse(t, recorder.Body, rspCandidate)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"name":      candidate.Name,
				"dob":       candidate.Dob,
				"bioLink":   candidate.BioLink,
				"imageLink": candidate.ImageUrl,
				"policy":    candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCandidate(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Candidate{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBioLink",
			body: gin.H{
				"name":      candidate.Name,
				"dob":       candidate.Dob,
				"bioLink":   "Invalid",
				"imageLink": candidate.ImageUrl,
				"policy":    candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidImageLink",
			body: gin.H{
				"name":      candidate.Name,
				"dob":       candidate.Dob,
				"bioLink":   candidate.BioLink,
				"imageLink": "Invalid",
				"policy":    candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidDob",
			body: gin.H{
				"name":      candidate.Name,
				"dob":       "Invalid",
				"bioLink":   candidate.BioLink,
				"imageLink": candidate.ImageUrl,
				"policy":    candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCandidate(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/candidates")
			values, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(values))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func TestGetCandidateAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	candidate := RandomCandidate()
	resultRow := db.GetCandidateRow{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
	}
	rspCandidate := NewCandidateResponse(candidate)

	testCases := []struct {
		name          string
		candidateID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(resultRow, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateResponse(t, recorder.Body, rspCandidate)
			},
		},
		{
			name:        "InvalidID",
			candidateID: -1,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "NotFound",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(db.GetCandidateRow{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(db.GetCandidateRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/api/candidates/%d", tc.candidateID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func TestListCandidatesAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	n := 5
	candidates := make([]db.Candidate, n)
	resultRows := make([]db.ListCandidatesRow, n)
	for i := 0; i < n; i++ {
		candidates[i] = RandomCandidate()
		resultRows[i] = db.ListCandidatesRow{
			ID:        candidates[i].ID,
			Name:      candidates[i].Name,
			Dob:       candidates[i].Dob,
			BioLink:   candidates[i].BioLink,
			ImageUrl:  candidates[i].ImageUrl,
			Policy:    candidates[i].Policy,
			VoteCount: candidates[i].VoteCount,
		}
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.ListCandidatesParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListCandidates(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(resultRows, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidates(t, recorder.Body, resultRows)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCandidates(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListCandidatesRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				pageID:   1,
				pageSize: 1000000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCandidates(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCandidates(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/candidates")

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func TestUpdateCandidateAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	candidate := RandomCandidate()
	resultRow := db.UpdateCandidateRow{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
	}
	rspCandidate := NewCandidateResponse(candidate)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     candidate.BioLink,
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.UpdateCandidateParams{
					ID:       candidate.ID,
					Name:     candidate.Name,
					Dob:      candidate.Dob,
					BioLink:  candidate.BioLink,
					ImageUrl: candidate.ImageUrl,
					Policy:   candidate.Policy,
				}

				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(resultRow, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateResponse(t, recorder.Body, rspCandidate)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     candidate.BioLink,
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateCandidateRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     candidate.BioLink,
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateCandidateRow{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InvalidDob",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         "Invalid",
				"bioLink":     candidate.BioLink,
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidBioLink",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     "Invalid",
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidImageLink",
			body: gin.H{
				"candidateId": candidate.ID,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     candidate.BioLink,
				"imageLink":   "Invalid",
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCandidateId",
			body: gin.H{
				"candidateId": -1,
				"name":        candidate.Name,
				"dob":         candidate.Dob,
				"bioLink":     candidate.BioLink,
				"imageLink":   candidate.ImageUrl,
				"policy":      candidate.Policy,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCandidate(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/candidates")
			values, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(values))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func TestDeleteCandidateAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	candidate := RandomCandidate()
	resultRow := db.GetCandidateRow{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
	}

	testCases := []struct {
		name          string
		candidateID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(resultRow, nil)
				store.EXPECT().
					DeleteCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "GetCandidateInternalError",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(db.GetCandidateRow{}, sql.ErrConnDone)
				store.EXPECT().
					DeleteCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "DeleteCandidateInternalError",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(resultRow, nil)
				store.EXPECT().
					DeleteCandidate(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "NotFound",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return(db.GetCandidateRow{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteCandidate(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:        "InvalidID",
			candidateID: -1,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCandidate(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteCandidate(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/candidates/%d", tc.candidateID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func requireBodyMatchCandidateResponse(t *testing.T, body *bytes.Buffer, candidate candidateResponse) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCandidate candidateResponse
	err = json.Unmarshal(data, &gotCandidate)
	require.NoError(t, err)
	require.Equal(t, candidate, gotCandidate)
}

func requireBodyMatchCandidate(t *testing.T, body *bytes.Buffer, candidate db.Candidate) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCandidate db.Candidate
	err = json.Unmarshal(data, &gotCandidate)
	require.NoError(t, err)
	require.Equal(t, candidate, gotCandidate)
}

func requireBodyMatchCandidates(t *testing.T, body *bytes.Buffer, candidates []db.ListCandidatesRow) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotCandidates []db.ListCandidatesRow
	err = json.Unmarshal(data, &gotCandidates)
	require.NoError(t, err)
	require.Equal(t, candidates, gotCandidates)
}

func RandomCandidate() db.Candidate {
	return db.Candidate{
		ID:         util.RandomInt(1, 1000),
		Name:       util.RandomName(),
		Dob:        util.RandomDob(),
		BioLink:    util.RandomBioLink(),
		ImageUrl:   util.RandomImageLink(),
		Policy:     util.RandomString(15),
		VoteCount:  0,
		Percentage: 0,
	}
}
func NewCandidateResponse(candidate db.Candidate) candidateResponse {
	return candidateResponse{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
	}
}
