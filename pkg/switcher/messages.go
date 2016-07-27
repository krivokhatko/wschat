package switcher

type ChatMessage struct {
	NickName string `json:"nickname"`
	Text     string `json:"text"`
}

type ChatMessages struct {
	Messages []*ChatMessage `json:"messages"`
}
