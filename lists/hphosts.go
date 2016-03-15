package lists

import "github.com/sprungknoedl/reputile/model"

var hphosts = List{
	Key:         "hosts-file.net",
	Name:        "hpHosts",
	URL:         "http://hosts-file.net/",
	Description: `hpHosts is a community managed and maintained hosts file that allows an additional layer of protection against access to ad, tracking and malicious websites.`,
	Iterator: func() Iterator {
		fn := func(category, description string) func(row []string) *model.Entry {
			return func(row []string) *model.Entry {
				if len(row) < 2 {
					return nil
				}

				if row[1] == "localhost" {
					return nil
				}

				return &model.Entry{
					Domain:      row[1],
					Category:    category,
					Description: description,
				}
			}
		}

		return Combine(
			SSV("http://hosts-file.net/emd.txt", fn("malware", "engaged in malware distribution")),
			SSV("http://hosts-file.net/exp.txt", fn("malware", "engaged in the housing, development or distribution of exploits")),
			SSV("http://hosts-file.net/psh.txt", fn("phishing", "engaged in phishing")),
		)
	}(),
}

func init() {
	Lists = append(Lists, hphosts)
}
