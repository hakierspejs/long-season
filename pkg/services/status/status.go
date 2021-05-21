package status

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/hakierspejs/long-season/pkg/services/macs"
	"github.com/hakierspejs/long-season/pkg/storage"
)

// Daemon is function to be runned in the background.
type Daemon func()

// NewDeamon returns channel for communicating with deamon and deamon
// to be run in the background in the separate gourtine.
func NewDaemon(ctx context.Context,
	iter storage.StatusIterator, counters storage.StatusTx,
) (chan<- []net.HardwareAddr, Daemon) {
	ch := make(chan []net.HardwareAddr)
	decrCh := make(chan net.HardwareAddr)

	daemon := func() {
		// Slice with newest mac addresse
		macs := macs.NewCounterSet(10)

		// TODO(thinkofher) make time period configurable
		ticker := time.NewTicker(time.Minute)

		for {
			select {
			case <-ctx.Done():
				break
			case newMacs := <-ch: // Update mac addresses
				log.Println("Received new macs")

				for _, newMac := range newMacs {
					macs.Incr(newMac)

					// Decrease after one minute
					go func(addr string) {
						<-time.After(time.Minute)
						decrCh <- net.HardwareAddr(addr)
					}(newMac.String())

				}

				log.Println("Updated macs")
			case <-ticker.C: // Update users every minute with newest mac addresses
				// Update online status for every user in db
				err := storage.UpdateStatuses(ctx, storage.UpdateStatusesArgs{
					Addresses: macs.Slice(),
					Iter:      iter,
					Counters:  counters,
				})
				if err != nil {
					log.Println("Failed to update statuses, reason:  ", err.Error())
					continue
				}
				log.Println("Succefully updated stauses.")
			case toDecr := <-decrCh:
				macs.Decr(toDecr)
			}
		}
	}

	return ch, daemon
}
