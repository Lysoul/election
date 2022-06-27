package api

import (
	"fmt"

	db "election/db/sqlc"
	"election/token"
	"election/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     util.Config
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("dateOfBirth", validDob)
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/api").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/candidates", server.createCandidate)
	authRoutes.GET("/candidates/:id", server.getCandidate)
	authRoutes.GET("/candidates", server.listCandidates)
	authRoutes.PUT("/candidates", server.updateCandidate)
	authRoutes.DELETE("/candidates/:id", server.deleteCandidate)

	authRoutes.POST("/vote", server.voteCandidate)
	authRoutes.POST("/vote/status", server.checkVoteStatus)

	authRoutes.POST("/election/toggle", server.toggleElection)
	authRoutes.GET("/election/result", server.electionResult)
	authRoutes.GET("/election/export", server.exportCSVElectionResult)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func successResponse() gin.H {
	return gin.H{
		"status": "ok",
	}
}

func errorResponse(err error) gin.H {
	return gin.H{
		"status": "error",
		"error":  err.Error(),
	}
}
