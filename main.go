package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/spf13/viper"
)

const helpMessage = `Usage: 
@<bot-nickname> vv <keywords>

A bot for searching 张维为 quote picture according to given keywords.

Description:
    vv <keywords>    Searches for an image based on the keywords and sends it to the group.

Example:
    @<bot-nickname> vv cute cat

Note:
    Longer keywords for better results!
`

func main() {
	viper.SetConfigFile("config.yml")
	viper.AutomaticEnv()
	viper.BindEnv("napcatAPIHost", "NAPCAT_API_HOST")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	app := &Application{
		Config: &Config{
			NapcatAccessToken: viper.GetString("napcatAccessToken"),
			QQNumber:          viper.GetString("QQ"),
			GoListenPort:      viper.GetString("goListenPort"),
			NapcatAPIPort:     viper.GetString("napcatAPIPort"),
			NapcatAPIHost:     viper.GetString("napcatAPIHost"),
		},
	}

	http.HandleFunc("/", app.messageHandler())

	log.Printf("Starting server for napcat messages on http://localhost:%s", app.Config.GoListenPort)
	if err := http.ListenAndServe(":"+app.Config.GoListenPort, nil); err != nil {
		log.Fatalf("Failed to start server: %s\n", err)
	}
}

// 处理来自 napcat 的消息
func (app *Application) messageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { // 只接受 POST 请求
			log.Printf("Received a %s request, but only POST is supported.", r.Method)
			return
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			log.Printf("Error reading request body: %v", err)
			return
		}
		defer r.Body.Close()

		// 格式化 JSON
		var result QQMessageEvent
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			log.Printf("Raw body: %s", string(body))
			return
		}

		go app.processMessage(result)
		w.WriteHeader(http.StatusOK)
	}
}

func (app *Application) processMessage(event QQMessageEvent) {
	// 提取有用信息用于回调
	groupID := event.GroupID
	msg := event.RawMessage

	// 如果消息为空，则不处理
	if groupID == 0 || msg == "" {
		return
	}

	atMsg := "[CQ:at,qq=" + app.Config.QQNumber + "]"

	// 检查是否是at机器人的消息
	if !strings.Contains(msg, atMsg) {
		return
	}

	if app.isHelpMessage(msg, atMsg) {
		app.sendTextMessage(groupID, helpMessage)
		return
	}

	responseURL := app.vvquest(msg)
	if responseURL != "" {
		log.Printf("Found image URL: %s", responseURL)
		app.SendPic(groupID, responseURL)
	} else {
		// 如果 vvquest 返回空，且用户提到了机器人并使用了 "vv "，说明关键词为空或未找到
		// isHelpMessage 在这里为 false, 我们需要检查是否包含 "vv "
		if strings.Contains(msg, "vv ") {
			app.sendTextMessage(groupID, helpMessage)
		}
	}
}

func (app *Application) vvquest(m string) string {
	vvIndex := strings.Index(m, "vv ")
	if vvIndex == -1 {
		return ""
	}
	keyword := strings.TrimSpace(m[vvIndex+3:])
	if keyword == "" {
		return ""
	}
	keyword = url.QueryEscape(keyword) // URL 编码关键词

	resp, err := http.Get("https://api.zvv.quest/search?q=" + keyword + "&n=1")
	if err != nil {
		log.Printf("Error making request to zvv.quest: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("zvv.quest returned non-200 status: %d", resp.StatusCode)
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading vv.quest response body: %v", err)
		return ""
	}
	var vresp VVResp
	if err := json.Unmarshal(body, &vresp); err != nil {
		log.Printf("Error parsing vv.quest JSON response: %v", err)
		return ""
	}

	if vresp.Code >= 400 || len(vresp.URLs) == 0 {
		log.Printf("vv.quest returned error code or no data: code=%d, msg=%s", vresp.Code, vresp.Msg)
		return ""
	}

	return vresp.URLs[0]
}

// 通用发送消息方法
func (app *Application) sendMessage(groupID int64, message []MsgComponent) {
	config := app.Config
	napcatAccessToken := config.NapcatAccessToken

	napcatURL := fmt.Sprintf("http://%s:%s/send_group_msg", config.NapcatAPIHost, config.NapcatAPIPort)

	payload := SendMsgPayload{
		GroupID: groupID,
		Message: message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
		return
	}
	err = app.Post(napcatURL, jsonPayload, napcatAccessToken)
	if err == nil {
		log.Printf("Successfully sent message to group %d.", groupID)
	}
}

func (app *Application) SendPic(groupID int64, URL string) {
	u, err := url.Parse(URL)
	imageName := "我们的网民有很多创意" // 默认值
	if err == nil {
		imageName = filepath.Base(u.Path)
	}

	message := []MsgComponent{
		{
			Type: "image",
			Data: map[string]any{
				"file":    URL,
				"summary": imageName,
			},
		},
	}
	app.sendMessage(groupID, message)
}

func (app *Application) sendTextMessage(groupID int64, text string) {
	message := []MsgComponent{{
		Type: "text",
		Data: map[string]any{"text": text},
	}}
	app.sendMessage(groupID, message)
}
