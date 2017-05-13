package lists

import "net"

var greensnow = List{
	Key:         "greensnow.co",
	Name:        "GreenSnow",
	URL:         "https://greensnow.co/",
	Description: `GreenSnow is a team consisting of the best specialists in computer security, we harvest a large number of IPs from different computers located around the world. GreenSnow is comparable with SpamHaus.org for attacks of any kind except for spam. Our list is updated automatically and you can withdraw at any time your IP address if it has been listed.`,
	Iterator: CSV(
		"http://blocklist.greensnow.co/greensnow.txt",
		func(row []string) *Entry {
			return &Entry{
				IP:       net.ParseIP(row[0]),
				Category: "attacker",
			}
		},
	),
}

func init() {
	Lists = append(Lists, greensnow)
}
