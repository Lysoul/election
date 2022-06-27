package api

import (
	"database/sql"
	"net/http"
	"time"

	db "election/db/sqlc"

	"github.com/gin-gonic/gin"
)

type createCandidateRequest struct {
	Name      string `json:"name" binding:"required"`
	Dob       string `json:"dob" binding:"required,dateOfBirth"`
	BioLink   string `json:"bioLink" binding:"required,url"`
	ImageLink string `json:"imageLink" binding:"required,url"`
	Policy    string `json:"policy" binding:"required"`
}

func (server Server) createCandidate(ctx *gin.Context) {
	var req createCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateCandidateParams{
		Name:       req.Name,
		Dob:        req.Dob,
		BioLink:    req.BioLink,
		ImageUrl:   req.ImageLink,
		Policy:     req.Policy,
		VoteCount:  0,
		Percentage: 0,
	}

	candidate, err := server.store.CreateCandidate(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := candidateResponse{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
		CreateAt:  candidate.CreateAt,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type getCandidateRequest struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

type candidateResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Dob       string    `json:"dob"`
	BioLink   string    `json:"bio_link"`
	ImageUrl  string    `json:"image_url"`
	Policy    string    `json:"policy"`
	VoteCount int32     `json:"vote_count"`
	CreateAt  time.Time `json:"create_at"`
}

func (server Server) getCandidate(ctx *gin.Context) {
	var req getCandidateRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	candidate, err := server.store.GetCandidate(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := candidateResponse{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
		CreateAt:  candidate.CreateAt,
	}

	ctx.JSON(http.StatusOK, rsp)

}

type listCandidateRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server Server) listCandidates(ctx *gin.Context) {
	var req listCandidateRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCandidatesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	candidates, err := server.store.ListCandidates(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, candidates)

}

type updateCandidateRequest struct {
	CandidateId int64  `json:"candidateId" binding:"required,min=1"`
	Name        string `json:"name" binding:"required"`
	Dob         string `json:"dob" binding:"required,dateOfBirth"`
	BioLink     string `json:"bioLink" binding:"required,url"`
	ImageLink   string `json:"imageLink" binding:"required,url"`
	Policy      string `json:"policy" binding:"required"`
}

func (server Server) updateCandidate(ctx *gin.Context) {
	var req updateCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateCandidateParams{
		ID:       req.CandidateId,
		Name:     req.Name,
		Dob:      req.Dob,
		BioLink:  req.BioLink,
		ImageUrl: req.ImageLink,
		Policy:   req.Policy,
	}

	candidate, err := server.store.UpdateCandidate(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := candidateResponse{
		ID:        candidate.ID,
		Name:      candidate.Name,
		Dob:       candidate.Dob,
		BioLink:   candidate.BioLink,
		ImageUrl:  candidate.ImageUrl,
		Policy:    candidate.Policy,
		VoteCount: candidate.VoteCount,
		CreateAt:  candidate.CreateAt,
	}

	ctx.JSON(http.StatusOK, rsp)
}

type deleteCandidateRequest struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (server Server) deleteCandidate(ctx *gin.Context) {
	var req deleteCandidateRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//TODO: replace with more efficient query for checking candidate exist e.g. select top 1
	candidate, err := server.store.GetCandidate(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = server.store.DeleteCandidate(ctx, candidate.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse())

}
