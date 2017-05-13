package lists

var phishtank = List{
	Key:         "phishtank.com",
	Name:        "PhishTank",
	URL:         "http://www.phishtank.com/",
	Description: `PhishTank is a free community site where anyone can submit, verify, track and share phishing data.`,
	Iterator: CSV(
		"http://data.phishtank.com/data/online-valid.csv",
		func(row []string) *Entry {
			if row[0] == "phish_id" {
				// header line
				return nil
			}

			return &Entry{
				Domain:      ExtractHost(row[1]),
				Category:    "phishing",
				Description: "Domain hosts web pages used for phishing",
			}
		}),
}

func init() {
	Lists = append(Lists, phishtank)
}
