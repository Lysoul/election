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

type VoteStatusResponse struct {
	Status bool `json:"status"`
}

func TestCheckVoteStatusAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)

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
				"nationalId": user.NationalID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVoteStatus(t, recorder.Body, VoteStatusResponse{
					Status: user.HasVoted,
				})
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"nationalId": user.NationalID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"nationalId": user.NationalID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InvalidNationalID",
			body: gin.H{
				"national_id": "invalid#1",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/vote/status")
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

type VotedResponse struct {
	Status string `json:"status"`
}

func TestVoteCandidateAPI(t *testing.T) {
	user, _ := CreateRandomUser(t)
	user2, _ := CreateRandomUser(t)
	user2.HasVoted = true

	candidate := RandomCandidate()
	closedElectionProperty := CreateClosedElectionProperty()
	closedElectionProperty2 := CreateClosedElectionProperty()
	closedElectionProperty2.Value = true

	voted := CreateVoted(user.NationalID, candidate.ID)

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
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(1).
					Return(closedElectionProperty, nil)

				arg := db.CreateVoteParams{
					VoteNationalID: user.NationalID,
					CandidateID:    candidate.ID,
				}
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(voted, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchResponseOK(t, recorder.Body)
			},
		},
		{
			name: "NationalIDNotFound",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(0)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "NationalIDInternalError",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(0)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "AlreadyVoted",
			body: gin.H{
				"nationalId":  user2.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user2.NationalID)).
					Times(1).
					Return(user2, nil)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(0)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetClosedElectionInternalError",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(1).
					Return(db.ElectionProperty{}, sql.ErrConnDone)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ClosedElection",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(1).
					Return(closedElectionProperty2, nil)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "CreateVoteInternalError",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.NationalID)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(1).
					Return(closedElectionProperty, nil)
				arg := db.CreateVoteParams{
					VoteNationalID: user.NationalID,
					CandidateID:    candidate.ID,
				}
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Vote{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidNationalID",
			body: gin.H{
				"nationalId":  "invalid#1",
				"candidateId": candidate.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(0)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCandidateID",
			body: gin.H{
				"nationalId":  user.NationalID,
				"candidateId": -1,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.NationalID, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetElectionProperty(gomock.Any(), gomock.Eq(util.ElectionClosed)).
					Times(0)
				store.EXPECT().
					CreateVote(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/vote")
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

func requireBodyMatchVoteStatus(t *testing.T, body *bytes.Buffer, voteStatus VoteStatusResponse) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotVoteStatus VoteStatusResponse
	err = json.Unmarshal(data, &gotVoteStatus)
	require.NoError(t, err)
	require.Equal(t, voteStatus, gotVoteStatus)
}

func requireBodyMatchResponseOK(t *testing.T, body *bytes.Buffer) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotVotedResponse VotedResponse
	err = json.Unmarshal(data, &gotVotedResponse)
	require.NoError(t, err)
	require.Equal(t, "ok", gotVotedResponse.Status)
}

func CreateClosedElectionProperty() db.ElectionProperty {
	return db.ElectionProperty{
		ID:    util.RandomInt(1, 1000),
		Name:  util.ElectionClosed,
		Value: false,
	}
}
func CreateVoted(nationalID string, candidateId int64) db.Vote {
	return db.Vote{
		ID:             util.RandomInt(1, 1000),
		VoteNationalID: nationalID,
		CandidateID:    candidateId,
	}
}
