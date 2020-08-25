package client

//notifyBuffLength is the buffer length for the NotifyChannelType channels of a cache.
const notifyBuffLength = 100

//NotifyChannelType identifies a notification channel for a specific event.
type NotifyChannelType int

//NotifyChannelType identifiers.
const (
	//Notification channel id for the creation of an Advertisement
	ChanAdvNew NotifyChannelType = iota
	//Notification channel id for the acceptance of an Advertisement
	ChanAdvAccepted
	//Notification channel id for the deletion of an Advertisement
	ChanAdvDeleted
	//Notification channel id for the revocation of the 'ACCEPTED' status of an Advertisement
	ChanAdvRevoked
)

var notifyChannelNames = []NotifyChannelType{
	ChanAdvNew,
	ChanAdvAccepted,
	ChanAdvDeleted,
	ChanAdvRevoked,
}
