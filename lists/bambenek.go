package lists

import (
	"net"
)

var bambenek = List{
	Key:         "bambenekconsulting.com",
	Name:        "Bambenek Consulting OSINT",
	URL:         "http://www.bambenekconsulting.com/",
	Description: `Bambenek Consulting is an IT consulting firm focused on cybersecurity and cybercrime. Every day, there is another story about another company having their banking accounts drained, someone having their identity stolen, or critical infrastructure being taken offline by hostile entities. Led by IT security expert, John Bambenek, we have the resources to bring to your business so you can be sure your organization and your customers’ data is safe. And when disaster does strike, you know you can count on us to be with you every step of the way as you recover from an incident.`,
	Iterator: CSV(
		"http://osint.bambenekconsulting.com/feeds/c2-ipmasterlist.txt",
		func(row []string) *Entry {
			return &Entry{
				IP:          net.ParseIP(row[0]),
				Category:    "malware",
				Description: row[1],
			}
		}),
}

func init() {
	Lists = append(Lists, bambenek)
}
