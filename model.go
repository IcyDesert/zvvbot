package main

type Application struct {
	Config *Config
}

type Config struct {
	NapcatAccessToken string
	GoListenPort      string
	NapcatAPIPort     string
	NapcatAPIHost     string
	QQNumber          string
}

// ===== OneBot 协议核心消息结构 =====
type MsgComponent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// 接收消息模型
type QQMessageEvent struct {
	Font          int            `json:"font"`
	GroupID       int64          `json:"group_id"`
	GroupName     string         `json:"group_name"`
	Message       []MsgComponent `json:"message"`
	MessageFormat string         `json:"message_format"`
	MessageID     int64          `json:"message_id"`
	MessageSeq    int64          `json:"message_seq"`
	MessageType   string         `json:"message_type"`
	PostType      string         `json:"post_type"`
	RawMessage    string         `json:"raw_message"`
	RealID        int64          `json:"real_id"`
	RealSeq       string         `json:"real_seq"`
	SelfID        int64          `json:"self_id"`
	Sender        SenderInfo     `json:"sender"`
	SubType       string         `json:"sub_type"`
	Time          int64          `json:"time"`
	UserID        int64          `json:"user_id"`
}

type SenderInfo struct {
	Card     string `json:"card"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	UserID   int64  `json:"user_id"`
}

// 发送消息模型
type SendMsgPayload struct {
	GroupID int64          `json:"group_id"`
	Message []MsgComponent `json:"message"`
}

type VVResp struct {
	Code int      `json:"code"`
	URLs []string `json:"data"`
	Msg  string   `json:"msg"`
}
