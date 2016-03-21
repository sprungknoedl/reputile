# About reputile
reputile is a platform that combines many different sources for reputation checks
into a single downloadable and queryable format.

reputile is foremost meant to be used as an API or datasource. The provided
frontend is only meant for exploration. The database can be retrieved in total or
filtered based on user defined criterias. See [Query][query] for more
details.

##Format
The reputile format strives to be in the most simple form possible to enable easy
usage in a variety of languages / tools. Because a lot of reputation sources 
allready use some form of CSV format, the reputile format is based on CSV (to be 
more specify RFC4180) as well.

The columns of the reputile format are:

	source,domain,ip4,last,category,description

The detailed description of each field is:

* **source**: The name of the upstream source for this row
* **domain**: The DNS domain name this row refers to. *Can be empty if not applicable*
* **ip4**: The IPv4 address this row refers to. *Can be empty if not applicable*
* **last**: A UNIX timestamp (UTC) when this row was last refreshed or checked
* **category**: A category identifier for this row. See [Categories][categories] for details.
* **description**: A description for humans why this row is listed. Can be empty if not provided by upstream.

##Query
The reputile database can be filtered with URL query parameters. Each
provided query paramter specifies a condition a row must satisfy.

When filtering against a timestamp field, the query is intepreted as *GREATER THAN*.

###Examples

	?domain=example.com&category=malware

This query returns all rows that match the domain *example.com* and the category *malware*.

	?category=spam

This query returns all rows that match the category *spam*.
