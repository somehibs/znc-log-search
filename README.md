# znc-log-search
golang + arangodb + sphinx search = znc log search engine

this log search engine was originally written in python.
it was a total mess, and this isn't much better.

it is, however, a little cheaper to execute and no longer requires rabbitmq and 3x python instances

the aim of this system is to merge indexing many channels over multiple user accounts for a single network
it uses sphinxsearch in place of elasticsearch due to the high ram demands of elasticsearch

to read any data for display, the original log files will need to be retained and not modified
as the search index retains file seek indexes.
