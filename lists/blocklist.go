package lists

import (
	"net"
)

var blocklist = List{
	Key:         "blocklist.de",
	Name:        "blocklist.de",
	URL:         "http://www.blocklist.de/en/index.html",
	Description: `blocklist.de is a free and voluntary service provided by a Fraud/Abuse-specialist, whose servers are often attacked on SSH-, Mail-Login-, FTP-, Webserver- and other services. The mission is to report all attacks to the abuse deparments of the infected PCs/servers to ensure that the responsible provider can inform the customer about the infection and disable them.`,
	Iterator: CSV(
		"http://lists.blocklist.de/lists/all.txt",
		func(row []string) *Entry {
			return &Entry{
				IP:       net.ParseIP(row[0]),
				Category: "attacker",
			}
		},
	),
}

func init() {
	Lists = append(Lists, blocklist)
}
