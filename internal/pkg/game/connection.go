package game

type Connection interface {
	RunReceive(senderID int) error
	RunSend() error
	Stop()

	GetSendChan() chan EventWrapper
	GetReceiveChan() chan EventWrapper
	GetStopSendChan() chan bool
	GetStopReceiveChan() chan bool
}
