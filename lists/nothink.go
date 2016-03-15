package lists

import "github.com/sprungknoedl/reputile/model"

var nothink = List{
	Key:         "nothink.org",
	Name:        "nothink.org",
	URL:         "http://www.nothink.org/index.php",
	Description: `Matteo Cantoni, ICT senior security analyst and penetration tester runs several honeypots to detect SNMP, SSH, DNS and telnet attacks`,
	Iterator: func() Iterator {
		fn := func(description string) func(row []string) *model.Entry {
			return func(row []string) *model.Entry {
				return &model.Entry{
					IP4:         row[0],
					Category:    "attacker",
					Description: description,
				}
			}
		}

		return Combine(
			CSV("http://www.nothink.org/blacklist/blacklist_snmp_week.txt", fn("SNMP attackers")),
			CSV("http://www.nothink.org/blacklist/blacklist_ssh_week.txt", fn("SSH attackers")),
			CSV("http://www.nothink.org/blacklist/blacklist_telnet_week.txt", fn("Telnet attackers")),
		)
	}(),
}

func init() {
	Lists = append(Lists, nothink)
}
