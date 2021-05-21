package status

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/hakierspejs/long-season/pkg/services/macs"
	"github.com/hakierspejs/long-season/pkg/storage"
)

// Daemon is a background function
type Daemon func()

// DaemonArgs contains list of arguments for NewDeamon constructor.
type DaemonArgs struct {
	Iter storage.StatusIterator

	Counters storage.StatusTx

	// RefreshTime is duration, that every time when passes, users
	// get their online status updated.
	RefreshTime time.Duration

	// SingleAddrTTL represents time to live for single
	// mac address. After this period of time user with
	// given mac address will be marked as offline
	// during next status update.
	SingleAddrTTL time.Duration
}

// NewDeamon returns channel for communicating with daemon and daemon
// to be run in the background in the separate gourtine .
func NewDaemon(ctx context.Context, args DaemonArgs) (chan<- []net.HardwareAddr, Daemon) {
	ch := make(chan []net.HardwareAddr)

	daemon := func() {
		macs := macs.NewSetTTL(ctx)

		// Update users every t, t = args.RefreshTime
		ticker := time.NewTicker(args.RefreshTime)

		for {
			select {
			case <-ctx.Done():
				break
			case newMacs := <-ch: // Update mac addresses
				log.Println("Received new macs")
				for _, newMac := range newMacs {
					macs.Push(newMac, args.SingleAddrTTL)
				}
			case <-ticker.C:
				// Update online status for every user in db
				err := storage.UpdateStatuses(ctx, storage.UpdateStatusesArgs{
					Addresses: macs.Slice(),
					Iter:      args.Iter,
					Counters:  args.Counters,
				})
				if err != nil {
					log.Println("Failed to update statuses, reason:  ", err.Error())
					continue
				}
				log.Println("Succefully updated stauses.")
			}
		}
	}

	return ch, daemon
}
