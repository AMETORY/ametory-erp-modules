package objects

type ChatMessage interface {
	SendChatMessage() (any, error)
}

func SendChatMessage(msg ChatMessage) (any, error) {
	return msg.SendChatMessage()
}

func SendChatMessages(msgs []ChatMessage) (any, error) {
	for _, msg := range msgs {
		_, err := msg.SendChatMessage()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
