package types

type CompletionRequest struct {
	Message string
}
type CompletionResponse struct {
	Message string
}

type SlackMessageEvent struct {
	Token    string `json:"token"`
	TeamId   string `json:"team_id"`
	ApiAppId string `json:"api_app_id"`
	Event    struct {
		Type        string `json:"type"`
		EventTs     string `json:"event_ts"`
		User        string `json:"user"`
		Text        string `json:"text"`
		Ts          string `json:"ts"`
		Channel     string `json:"channel"`
		ChannelType string `json:"channel_type"`
	} `json:"event"`
	Type        string   `json:"type"`
	EventId     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
}
