package lists

var cybercrime = List{
	Key:         "cybercrime-tracker.net",
	Name:        "CyberCrime Tracker",
	URL:         "http://cybercrime-tracker.net/",
	Description: `CyberCrime tracks C&C servers`,
	Iterator: CSV(
		"http://cybercrime-tracker.net/all.php",
		func(row []string) *Entry {
			host := ExtractHost(row[0])
			return &Entry{
				Domain:   host,
				Category: "malware",
			}
		},
	),
}

func init() {
	Lists = append(Lists, cybercrime)
}
