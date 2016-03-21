package lists

import (
	"net"

	"github.com/sprungknoedl/reputile/model"
)

var abusech = List{
	Key:         "abuse.ch",
	Name:        "abuse.ch - ZeuS, Feodo & Ransomware Tracker",
	URL:         "https://www.abuse.ch/",
	Description: `abuse.ch tracks Command&Control servers for ZeuS and Feodo trojan and Ransomware around the world and provides you a domain- and a IP-blocklist.`,
	Iterator: func() Iterator {
		domain := func(description string) func(row []string) *model.Entry {
			return func(row []string) *model.Entry {
				return &model.Entry{
					Domain:      row[0],
					Category:    "malware",
					Description: description,
				}
			}
		}
		ip := func(description string) func(row []string) *model.Entry {
			return func(row []string) *model.Entry {
				return &model.Entry{
					IP:          net.ParseIP(row[0]),
					Category:    "malware",
					Description: description,
				}
			}
		}
		return Combine(
			CSV("https://zeustracker.abuse.ch/blocklist.php?download=baddomains", domain("ZeuS C&C server")),
			CSV("https://zeustracker.abuse.ch/blocklist.php?download=badips", ip("ZeuS C&C server")),
			CSV("https://feodotracker.abuse.ch/blocklist/?download=domainblocklist", domain("Feodo trojan C&C server")),
			CSV("https://feodotracker.abuse.ch/blocklist/?download=badips", ip("Feodo trojan C&C server")),
			CSV("https://ransomwaretracker.abuse.ch/downloads/RW_DOMBL.txt", domain("Ransomware botnet C&C traffic")),
			CSV("https://ransomwaretracker.abuse.ch/downloads/RW_IPBL.txt", ip("Ransomware botnet C&C traffic")),
			CSV("https://palevotracker.abuse.ch/blocklists.php?download=domainblocklist", domain("Palevo botnet C&C traffic")),
			CSV("https://palevotracker.abuse.ch/blocklists.php?download=ipblocklist", ip("Palevo botnet C&C traffic")),
		)
	}(),
}

func init() {
	Lists = append(Lists, abusech)
}
