package lists

import (
	"net"

	"github.com/sprungknoedl/reputile/model"
)

var bruteforceblocker = List{
	Key:         "rulez.sk",
	Name:        "BruteForceBlocker",
	URL:         "http://danger.rulez.sk/index.php/bruteforceblocker/",
	Description: `BruteForceBlocker is a perl script, that works along with pf. Its main purpose is to block SSH bruteforce attacks via firewall. When this script is running, it checks sshd logs from syslog and looks for failed login attempts and counts the number of such attempts. When a given IP reaches the configured limit of fails, the script puts this IP to the pf’s table and block any further traffic to the that box from the given IP.`,
	Iterator: TSV(
		"http://danger.rulez.sk/projects/bruteforceblocker/blist.php",
		func(row []string) *model.Entry {
			return &model.Entry{
				IP:       net.ParseIP(row[0]),
				Category: "attacker",
			}
		},
	),
}

func init() {
	Lists = append(Lists, bruteforceblocker)
}
