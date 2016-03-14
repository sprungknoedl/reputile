package lists

import "github.com/sprungknoedl/reputile/model"

var badips = List{
	Key:         "badips.com",
	Name:        "badips.com",
	URL:         "https://www.badips.com/",
	Description: `badips.com is a community based IP blacklist service. You can report malicious IPs to badips.com and you can download blacklists or query their API to find out if a IP is listed.`,
	Generator: func() Generator {
		fn := func(category, description string) func(row []string) *model.Entry {
			return func(row []string) *model.Entry {
				return &model.Entry{
					Source:      "badips.com",
					IP4:         row[0],
					Category:    category,
					Description: description,
				}
			}
		}

		return Combine(
			CSV("https://www.badips.com/get/list/ssh/3", fn("attacker", "SSH bruteforce login attacks and other ssh related attacks")),
			CSV("https://www.badips.com/get/list/dns/3", fn("attacker", "Attacks against the Domain Name System")),
			CSV("https://www.badips.com/get/list/http/3", fn("attacker", "Attacks aiming at HTTP/S services")),
		)
	}(),
}

func init() {
	Lists = append(Lists, badips)
}
