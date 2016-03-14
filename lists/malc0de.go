package lists

import "github.com/sprungknoedl/reputile/model"

var malc0de = List{
	Key:         "malc0de.com",
	Name:        "malc0de",
	URL:         "http://malc0de.com/dashboard/",
	Description: `malc0de is an updated database of domains hosting malicious executables during the last 30 days.`,
	Generator: Combine(
		CStyleSSV("http://malc0de.com/bl/BOOT",
			func(row []string) *model.Entry {
				if len(row) < 2 {
					return nil
				}

				return &model.Entry{
					Source:      "malc0de.com",
					Domain:      row[1],
					Category:    "malware",
					Description: "distributed malware in the last 30 days",
				}
			}),
		CStyleCSV("http://malc0de.com/bl/IP_Blacklist.txt",
			func(row []string) *model.Entry {
				if len(row) < 1 {
					return nil
				}

				return &model.Entry{
					Source:      "malc0de.com",
					IP4:         row[0],
					Category:    "malware",
					Description: "distributed malware in the last 30 days",
				}
			})),
}

func init() {
	Lists = append(Lists, malc0de)
}
