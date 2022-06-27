package api

import (
	"database/sql"
	"net/http"
	"time"

	db "election/db/sqlc"
	"election/util"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createUserRequest struct {
	NationalID string `json:"national_id" binding:"required,number,len=13"`
	Password   string `json:"password" binding:"required,min=6"`
	Fullname   string `json:"full_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
}

type createUserResponse struct {
	NationalID        string    `json:"national_id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	Permission        []string  `json:"permission"`
	HasVoted          bool      `json:"has_voted"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreateAt          time.Time `json:"create_at"`
}

type userResponse struct {
	NationalID        string    `json:"national_id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		NationalID:        user.NationalID,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreateAt,
	}
}

func (server Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	permission := make([]string, 1)
	permission[0] = util.Vote

	arg := db.CreateUserParams{
		NationalID:     req.NationalID,
		HashedPassword: hashedPassword,
		FullName:       req.Fullname,
		Email:          req.Email,
		Permission:     permission,
		HasVoted:       false,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := createUserResponse{
		NationalID:        user.NationalID,
		FullName:          user.FullName,
		Email:             user.Email,
		Permission:        user.Permission,
		HasVoted:          user.HasVoted,
		PasswordChangedAt: user.PasswordChangedAt,
		CreateAt:          user.CreateAt,
	}

	ctx.JSON(http.StatusOK, response)
}

type loginUserRequest struct {
	NationalID string `json:"national_id" binding:"required,number,len=13"`
	Password   string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiredAt   time.Time    `json:"expired_at"`
	User        userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.NationalID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.NationalID,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		ExpiredAt:   accessPayload.ExpiredAt,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}
