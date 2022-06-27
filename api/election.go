package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"

	db "election/db/sqlc"
	"election/util"

	"github.com/gin-gonic/gin"
)

type toggleElectionRequest struct {
	Enable bool `json:"enable"`
}

func (server Server) toggleElection(ctx *gin.Context) {
	var req toggleElectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateElectionPropertyParams{
		Name:  util.ElectionClosed,
		Value: !req.Enable,
	}

	electionProperty, err := server.store.UpdateElectionProperty(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"enable": !electionProperty.Value,
	})
}

func (server Server) electionResult(ctx *gin.Context) {

	electionResults, err := server.store.ListCandidatesResult(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, electionResults)
}

func (server Server) exportCSVElectionResult(ctx *gin.Context) {

	FileName := "export.csv"
	votedLists, err := server.store.ListVoteOrderByCandidate(ctx)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	if err := w.Write([]string{"Candidate id", "National id"}); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	for _, votedList := range votedLists {
		var record []string
		record = append(record, strconv.FormatInt(votedList.CandidateID, 10))
		record = append(record, votedList.VoteNationalID)
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", FileName))
	ctx.Data(http.StatusOK, "text/csv", b.Bytes())
}
