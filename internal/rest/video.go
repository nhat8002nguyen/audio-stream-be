package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/nhat8002nguyen/audio-stream-be/domain"
)

// VideoService represent the video's usecases
//
//go:generate mockery --name VideoService
type VideoService interface {
	SearchVideos(ctx context.Context, text string, amount int64) ([]domain.SearchedVideo, error)
	GetStreamReader(ctx context.Context) (io.ReadCloser, error)
}

// VideoHandler  represent the httphandler for video
type VideoHandler struct {
	Service VideoService
}

// NewVideoHandler will initialize the videos/ resources endpoint
func NewVideoHandler(e *echo.Echo, svc VideoService) {
	handler := &VideoHandler{
		Service: svc,
	}
	e.GET("/videos/search", handler.SearchVideos)
	e.GET("/videos/ws", handler.HandleStreamWs)
}

// FetchVideo will fetch the video based on given params
func (a *VideoHandler) SearchVideos(c echo.Context) error {

	search := c.QueryParam("text")
	amount, err := strconv.ParseInt(c.QueryParam("amount"), 10, 64)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	ctx := c.Request().Context()
	searchedMetas, err := a.Service.SearchVideos(ctx, search, int64(amount))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	data := make(map[string]any)
	data["total"] = len(searchedMetas)
	data["data"] = searchedMetas

	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Replace this with your origin validation logic
		// (e.g., check allowed origins or perform authentication)
		return true
	},
}

func (a *VideoHandler) HandleStreamWs(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	defer conn.Close()

	ctx := c.Request().Context()
	streamReader, err := a.Service.GetStreamReader(ctx)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		// Process the received message
		fmt.Printf("Received message: %s\n", message)

		chunk := make([]byte, 1024)
		_, err = streamReader.Read(chunk)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
			}
		}

		// Optionally, send a response message
		err = conn.WriteMessage(messageType, chunk)
		if err != nil {
			return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
		}
	}

	return c.JSON(http.StatusOK, nil)
}
