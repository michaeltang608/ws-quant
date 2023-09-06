package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
	"strconv"
	"strings"
	"time"
	"ws-quant/cex"
	"ws-quant/common/bean"
	"ws-quant/common/symb"
)

func (s *Service) connectAndSubscribePublicBit() {

	s.connectPublic()
	s.subscribeTickers()

}

func (s *Service) subscribeTickersBit() {
	var err error
	argList := make([]map[string]interface{}, 0)
	for _, symbol_ := range symb.GetAllOkFuture() {

		arg := make(map[string]interface{})
		arg["channel"] = "tickers"
		arg["instId"] = fmt.Sprintf("%s-USDT", strings.ToUpper(symbol_))
		argList = append(argList, arg)
	}

	req := &Req{
		Op:   "subscribe",
		Args: argList,
	}
	reqBytes, _ := json.Marshal(req)
	err = s.pubConBit.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Panic("发送OKEX订阅消息失败 ", err)
	}
	log.Info("订阅全部tickers数据成功")

}
func (s *Service) connectPublicBit() {
	// 可能会重连
	log.Info("开始连接pub con")
	var err error
	socketUrl := "wss://ws.okx.com:8443/ws/v5/public"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil || conn == nil {
		// 第二次尝试连接，提高胜率
		conn, _, err = websocket.DefaultDialer.Dial(socketUrl, nil)
		if err != nil || conn == nil {
			log.Panic("service socket 连续两次连接失败", err.Error())
		}
	}
	if conn == nil {
		log.Info("奇怪，conn 还是 null")
	}
	s.pubConBit = conn
	s.pubConLastConnectTime = time.Now().Second()
	log.Info("连接pubCon 成功，开始监听消息了, pubCon==nil, %v", s.pubConBit == nil)
}

func (s *Service) listenAndNotifyPublicBit() {
	errCnt := 0
	for {
		if s.pubConBit == nil {
			time.Sleep(time.Second)
			continue
		}
		_, msgBytes, err := s.pubConBit.ReadMessage()
		if err != nil {
			log.Error("Error in receive:", err)
			time.Sleep(time.Second)
			errCnt++
			if errCnt > 10 {
				log.Info("读取失败累计超过10次，开始重启")
				log.Panic("service read pub err")
			}
			continue
		}

		errCnt = 0
		/*
			接受到的数据有如下几种场景
			1 接受到 pong
			2 接受到event
				- 如果是login, 立刻订阅
		*/
		msg := string(msgBytes)
		if msg == "pong" {
			log.Info("获取pong数据")
			continue
		}

		//_, ok := resp["data"]
		if fastjson.GetString(msgBytes, "arg", "channel") == "tickers" {
			// 收到 ticker 数据
			// 2 获取价格
			bestAsk := fastjson.GetString(msgBytes, "data", "0", "askPx")
			price := fastjson.GetString(msgBytes, "data", "0", "last")
			bestBid := fastjson.GetString(msgBytes, "data", "0", "bidPx")

			instId := fastjson.GetString(msgBytes, "data", "0", "instId")
			symbolStr := strings.Split(instId, "-")[0]
			priceBeatAskFloat, _ := strconv.ParseFloat(bestAsk, 64)
			priceFloat, _ := strconv.ParseFloat(price, 64)
			priceBestBidFloat, _ := strconv.ParseFloat(bestBid, 64)

			for _, symbol_ := range symb.GetAllOkFuture() {
				symbol := strings.ToUpper(symbol_)
				if symbol == strings.ToUpper(symbolStr) {
					tickerBean := bean.TickerBean{
						CexName:      cex.OKE,
						InstId:       instId,
						SymbolName:   symbol,
						Price:        priceFloat,
						PriceBestBid: priceBestBidFloat,
						PriceBestAsk: priceBeatAskFloat,
						Ts0:          time.Now().UnixMilli(),
					}
					s.tickerChan <- tickerBean
				}
			}

		} else {
			if fastjson.GetString(msgBytes, "event") == "subscribe" {
				//ignore
			} else if fastjson.GetString(msgBytes, "event") == "error" {
				log.Info("public 接收到订阅失败事件：%v\n", string(msgBytes))
			} else {
				log.Info("public 接收到未知业务数据：%v\n", string(msgBytes))
			}
		}
	}
}
