package validator

type ChatRequest struct {
	MessageText    string `json:"message_text"`
	Type           uint   `json:"type"`
	ImageUrl       string `json:"image_url"`
	SenderUserId   string `json:"sender_user_id"`
	ReceiverUserId string `json:"receiver_user_id"`
}
