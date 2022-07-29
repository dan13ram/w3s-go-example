package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/web3-storage/go-w3s-client"
)

var storage w3s.Client

type ServerResponse struct {
	Response interface{} `json:"response"`
	Error    string      `json:"error"`
}

func getStatus(ctx *gin.Context) {
	resp := ServerResponse{
		Response: "Welcome to a sample server!",
		Error:    "",
	}
	ctx.IndentedJSON(http.StatusOK, resp)
}

func uploadJSON(ctx *gin.Context) {
	jsonMap := make(map[string](interface{}))
	if err := ctx.BindJSON(&jsonMap); err != nil {
		resp := ServerResponse{
			Response: "Got Invalid JSON!",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusBadRequest, resp)
		return
	}
	fmt.Printf("INFO: jsonMap, %s\n", jsonMap)

	f, err := os.CreateTemp("", "tmp-json-")
	if err != nil {
		resp := ServerResponse{
			Response: "Internal Server Error",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusInternalServerError, resp)
	}

	defer f.Close()
	defer os.Remove(f.Name())

	json, err := json.Marshal(jsonMap)
	if err != nil {
		resp := ServerResponse{
			Response: "Internal Server Error",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusInternalServerError, resp)
	}

	data := []byte(json)

	if _, err := f.Write(data); err != nil {
		resp := ServerResponse{
			Response: "Internal Server Error",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusInternalServerError, resp)
	}

	readFile, err := os.Open(f.Name())
	if err != nil {
		resp := ServerResponse{
			Response: "Internal Server Error",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusInternalServerError, resp)
	}

	bgContext := context.Background()

	cid, err := storage.Put(bgContext, readFile)
	if err != nil {
		resp := ServerResponse{
			Response: "Internal Server Error",
			Error:    err.Error(),
		}
		ctx.IndentedJSON(http.StatusInternalServerError, resp)
	}

	resp := ServerResponse{
		Response: cid.String(),
		Error:    "",
	}
	ctx.IndentedJSON(http.StatusOK, resp)
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	fmt.Println("Starting server!")
	W3S_TOKEN := os.Getenv("WEB3_STORAGE_TOKEN")
	if W3S_TOKEN == "" {
		log.Fatalf("No env for WEB3_STORAGE_TOKEN found")
	}
	storage, err = w3s.NewClient(w3s.WithToken(W3S_TOKEN))
	if err != nil {
		log.Fatalf("Could not create storage client")
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/", getStatus)
	router.GET("/status", getStatus)
	router.POST("/json", uploadJSON)

	router.Run("localhost:9090")
}
