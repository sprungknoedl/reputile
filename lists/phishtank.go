package lists

import "github.com/sprungknoedl/reputile/model"

var phishtank = List{
	Key:         "phishtank.com",
	Name:        "PhishTank",
	URL:         "http://www.phishtank.com/",
	Description: `PhishTank is a free community site where anyone can submit, verify, track and share phishing data.`,
	Generator: CSV(
		"http://data.phishtank.com/data/online-valid.csv",
		func(row []string) *model.Entry {
			if row[0] == "phish_id" {
				// header line
				return nil
			}

			domain := ExtractHost(row[1])
			if domain == "" {
				return nil
			}

			return &model.Entry{
				Source:      "phishtank.com",
				Domain:      domain,
				Category:    "phishing",
				Description: "Domain hosts web pages used for phishing",
			}
		}),
}

func init() {
	Lists = append(Lists, phishtank)
}
