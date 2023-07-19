package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itoqsky/InnoCoTravel-backend/internal/core"
	"github.com/itoqsky/InnoCoTravel-backend/internal/server"
	"github.com/itoqsky/InnoCoTravel-backend/pkg/response"
)

func (h *Handler) initWsRoutes(api *gin.RouterGroup) {
	ws := api.Group("/ws", h.userIdentity)
	{
		ws.GET("/join-trip/:id", h.joinTrip)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) joinTrip(c *gin.Context) {
	uctx, err := getUserCtx(c)
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tripId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// userId, err := strconv.ParseInt(c.Query("userId"), 10, 64)
	// if err != nil {
	// 	response.NewErrorResponse(c, http.StatusBadRequest, err.Error())
	// 	return
	// }
	// username := c.Query("username")
	// if err != nil {
	// 	response.NewErrorResponse(c, http.StatusBadRequest, err.Error())
	// 	return
	// }

	_, err = h.services.Trip.GetById(uctx.UserId, tripId)
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cl := &server.Client{
		Conn:     conn,
		Message:  make(chan *core.Message, 10),
		Id:       uctx.UserId,
		Username: uctx.Username,
		RoomId:   tripId,
	}

	m := &core.Message{
		FromUsername: uctx.Username,
		FromUserId:   uctx.UserId,
		ToRoomId:     tripId,
		Content:      uctx.Username + " joined the room",
		ContentType:  core.TEXT,
		Url:          "",
	}

	h.hub.Register <- cl
	h.hub.Broadcast <- m

	go cl.WriteMessage()
	cl.ReadMessage(h.hub)
}
