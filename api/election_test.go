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

type ToggleElectionResponse struct {
	Enable bool   `json:"enable"`
	Status string `json:"status"`
}

func TestToggleElectionAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	enable := true
	electionClosed := db.ElectionProperty{
		ID:    util.RandomInt(1, 1000),
		Name:  util.ElectionClosed,
		Value: !enable,
	}

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
				"enable": enable,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.UpdateElectionPropertyParams{
					Name:  util.ElectionClosed,
					Value: !enable,
				}
				store.EXPECT().
					UpdateElectionProperty(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(electionClosed, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"enable": enable,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateElectionProperty(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ElectionProperty{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/election/toggle")
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

func TestGetElectionResultAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	n := 1
	candidates := make([]db.Candidate, n)
	resultRows := make([]db.ListCandidatesResultRow, n)
	for i := 0; i < n; i++ {
		candidates[i] = RandomCandidate()
		pst := fmt.Sprintf("%d", candidates[i].Percentage) + "%"
		resultRows[i] = db.ListCandidatesResultRow{
			ID:         candidates[i].ID,
			Name:       candidates[i].Name,
			Dob:        candidates[i].Dob,
			BioLink:    candidates[i].BioLink,
			ImageUrl:   candidates[i].ImageUrl,
			Policy:     candidates[i].Policy,
			VoteCount:  candidates[i].VoteCount,
			Percentage: pst,
		}
	}

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCandidatesResult(gomock.Any()).
					Times(1).
					Return(resultRows, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchElectionResult(t, recorder.Body, resultRows)
			},
		},
		{
			name: "InternalError",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListCandidatesResult(gomock.Any()).
					Times(1).
					Return([]db.ListCandidatesResultRow{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/election/result")
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func TestExportCSVElectionResultAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	n := 2
	candidate := RandomCandidate()
	resultRows := make([]db.ListVoteOrderByCandidateRow, n)
	for i := 0; i < n; i++ {
		resultRows[i] = db.ListVoteOrderByCandidateRow{
			CandidateID:    candidate.ID,
			VoteNationalID: user.NationalID,
		}
	}

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListVoteOrderByCandidate(gomock.Any()).
					Times(1).
					Return(resultRows, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListVoteOrderByCandidate(gomock.Any()).
					Times(1).
					Return([]db.ListVoteOrderByCandidateRow{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/api/election/export")
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)

		})

	}

}

func requireBodyMatchElectionResult(t *testing.T, body *bytes.Buffer, electionResult []db.ListCandidatesResultRow) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotElectionResult []db.ListCandidatesResultRow
	err = json.Unmarshal(data, &gotElectionResult)
	require.NoError(t, err)
	require.Equal(t, electionResult, gotElectionResult)
}
