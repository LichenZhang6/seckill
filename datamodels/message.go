package datamodels

// 简单的消息体
type Message struct {
	ProductID int64
	UserID    int64
}

// 创建结构体
func NewMessage(userID, productID int64) *Message {
	return &Message{UserID: userID, ProductID: productID}
}
