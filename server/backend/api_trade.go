package backend

import (
	"github.com/gin-gonic/gin"
	"ws-quant/core"
	"ws-quant/pkg/gintool"
)

func (bs *backendServer) openStrategy(cxt *gin.Context) {
	var req core.OrderReq
	err := cxt.Bind(&req)
	if err != nil {
		gintool.Error(cxt, err)
		return
	}
	msg := bs.okeService.StrategyOpenLimit(req.InstId, req.Price, req.Size, req.Side, "open")
	gintool.SucMsg(cxt, msg)
	return
}

func (bs *backendServer) openLimit(cxt *gin.Context) {
	var req core.OrderReq
	err := cxt.Bind(&req)
	if err != nil {
		gintool.Error(cxt, err)
		return
	}
	msg := bs.okeService.TradeLimit(req.InstId, req.Price, req.Size, req.Side, "open")
	gintool.SucMsg(cxt, msg)
	return
}

func (bs *backendServer) queryLiveOrder(cxt *gin.Context) {
	instId := cxt.Query("instId")
	gintool.SucMsg(cxt, bs.okeService.QueryLiveOrder(instId))
	return
}
func (bs *backendServer) cancelOrder(cxt *gin.Context) {

	service := bs.okeService
	msg := service.CancelOrder(cxt.Query("instId"), cxt.Query("orderId"))
	gintool.SucMsg(cxt, msg)
	return
}
func (bs *backendServer) closeMarket(cxt *gin.Context) {
	gintool.SucMsg(cxt, bs.okeService.CloseOrder(cxt.Query("orderType")))
	return
}