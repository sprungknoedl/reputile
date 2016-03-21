package lists

import "github.com/sprungknoedl/reputile/model"

var cybercrime = List{
	Key:         "cybercrime-tracker.net",
	Name:        "CyberCrime Tracker",
	URL:         "http://cybercrime-tracker.net/",
	Description: `CyberCrime tracks C&C servers`,
	Iterator: CSV(
		"http://cybercrime-tracker.net/all.php",
		func(row []string) *model.Entry {
			host := ExtractHost(row[0])
			return &model.Entry{
				Domain:   host,
				Category: "malware",
			}
		},
	),
}

func init() {
	Lists = append(Lists, cybercrime)
}
