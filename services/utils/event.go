package utils

import (
	"log"

	"github.com/diginatu/nagome/api"
)

// EmitEvNewNotification emits new event for ask UI to display a notification.
func EmitEvNewNotification(typ, title, desc string, evch chan<- *api.Message, log *log.Logger) {
	log.Printf("[D] %s : %s", title, desc)
	evch <- api.NewMessageMust(api.DomainUI, api.CommUINotification, api.CtUINotification{typ, title, desc})
}
