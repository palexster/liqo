package client

//notifyBuffLength is the buffer length for the NotifyChannel channels of a cache.
const notifyBuffLength = 100

//NotifyChannel identifies a notification channel for a specific event.
type NotifyChannel int

//NotifyChannel identifiers.
const (
	//Notification channel id for the creation of an Advertisement
	ChanAdvNew NotifyChannel = iota
	//Notification channel id for the acceptance of an Advertisement
	ChanAdvAccepted
	//Notification channel id for the deletion of an Advertisement
	ChanAdvDeleted
	//Notification channel id for the revocation of the 'ACCEPTED' status of an Advertisement
	ChanAdvRevoked
)

//notifyChannelNames contains all the registered NotifyChannel managed by the AgentController.
//It is used for init and testing purposes.
var notifyChannelNames = []NotifyChannel{
	ChanAdvNew,
	ChanAdvAccepted,
	ChanAdvDeleted,
	ChanAdvRevoked,
}
