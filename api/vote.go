package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "election/db/sqlc"
	"election/token"
	"election/util"

	"github.com/gin-gonic/gin"
)

var (
	ErrAlreadyVoted           = errors.New("Already voted")
	ErrClosedElection         = errors.New("Election is closed")
	ErrNoPermissionNationalID = errors.New("Cannot vote by another natinal ID")
)

type checkVoteStatusRequest struct {
	NationalId string `json:"nationalId" binding:"required,number,len=13"`
}

func (server Server) checkVoteStatus(ctx *gin.Context) {
	var req checkVoteStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.NationalId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": user.HasVoted})
}

type voteCandidateRequest struct {
	NationalId  string `json:"nationalId" binding:"required,number,len=13"`
	CandidateId int64  `json:"candidateId" binding:"required,min=1"`
}

func (server Server) voteCandidate(ctx *gin.Context) {
	var req voteCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.NationalId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if user.NationalID != authPayload.NationalID {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrNoPermissionNationalID))
		return
	}

	if user.HasVoted {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrAlreadyVoted))
		return
	}

	isClosedElection, err := server.store.GetElectionProperty(ctx, util.ElectionClosed)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if isClosedElection.Value {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrClosedElection))
		return
	}

	arg := db.CreateVoteParams{
		VoteNationalID: req.NationalId,
		CandidateID:    req.CandidateId,
	}

	_, err = server.store.CreateVote(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse())
}
