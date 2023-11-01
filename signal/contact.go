package signal

import "strconv"

func (a *Account) sendSyncContact(account uint64, pn []uint64) {
	//sn := []string{"+8617817059494", "+8618620027590", "+8619032288392", "+8619083218722", "+18722106579", "+234346645"}

	// Iterate through each phone number in the array

	for _, phoneNumber := range pn {
		// Create a sync contact request for the current phone number
		useJid := strconv.FormatUint(account, 10) + "@s.whatsapp.net"
		msg := a.signal.SyncContactRequest(useJid, []uint64{phoneNumber}, "1689755731-54726411-1")
		// Send the sync contact request for the current phone number
		a.send(msg)
	}
}
