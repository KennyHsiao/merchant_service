package types

type LineSendRequest struct {
	ChatID  string `json:"chatId, optional"`
	Message string `json:"message"`
}

type LineSendResponse struct {
	Msg string `json:"msg"`
}

type LineCallbackRequest struct {
	ChatID  string `json:"chatId"`
	Message string `json:"message"`
}

type LineCallbackResponse struct {
	Msg string `json:"msg"`
}

type TelegramSendRequest struct {
	ChatID  string `json:"chatId"`
	Message string `json:"message"`
}

type TelegramSendResponse struct {
	Msg string `json:"msg"`
}

type TelegramNotifyRequest struct {
	Message string `json:"message"`
}

type SlackSendRequest struct {
	ChatID  string `json:"chatId"`
	Message string `json:"message"`
}

type SlackSendResponse struct {
	Msg string `json:"msg"`
}
