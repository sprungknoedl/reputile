package lists

import (
	"net"

	"github.com/sprungknoedl/env"
	"github.com/sprungknoedl/reputile/model"
)

var autoshun = List{
	Key:         "autoshun.org",
	Name:        "AutoShun",
	URL:         "https://www.autoshun.org/",
	Description: `AutoShun is a Snort plugin that allows you to send your Snort IDS logs to a centralized server that will correlate attacks from your sensor logs with other snort sensors, honeypots, and mail filters from around the world.`,
	Iterator: CSV(
		"https://www.autoshun.org/download/?format=csv&api_key="+env.GetString("autoshun.apikey"),
		func(row []string) *model.Entry {
			if len(row) < 3 {
				return nil
			}

			return &model.Entry{
				IP:          net.ParseIP(row[0]),
				Category:    "attacker",
				Description: row[2],
			}
		}),
}

func init() {
	Lists = append(Lists, autoshun)
}
