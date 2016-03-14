package lists

import "github.com/sprungknoedl/reputile/model"

var cinsscore = List{
	Key:         "cinsscore.com",
	Name:        "CI Army List",
	URL:         "http://cinsscore.com/",
	Description: `The CI Army list is a subset of the CINS Active Threat Intelligence ruleset, and consists of IP addresses that meet two basic criteria: 1) The IP's recent Rogue Packet score factor is very poor, and 2) The InfoSec community has not yet identified the IP as malicious.`,
	Generator: CSV(
		"http://cinsscore.com/list/ci-badguys.txt",
		func(row []string) *model.Entry {
			return &model.Entry{
				Source:      "cinsscore.com",
				IP4:         row[0],
				Category:    "malware",
				Description: "",
			}
		},
	),
}

func init() {
	Lists = append(Lists, cinsscore)
}
