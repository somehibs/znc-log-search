# znc-log-search
golang + arangodb + sphinx search = znc log search engine

this log search engine was originally written in python with rabbitmq and 3 python instances for 'performance'.
it was a total mess, and this isn't much better.

sphinxsearch was chosen due to ram constraints

to read any data for display, the original log files will need to be retained and not modified, the search index only retains file seek indexes.

config object description is in conf.go, config.json is your custom json file

## go
golang.org has a download link, go get github.com/somehibs/znc-log-search

## arangodb
create a server, user, password and db for this application. configure them in the config. create collections Nicks, Channels and Users. the application will blow up until they're accessible.

## sphinx config
```index irc_msg
{
  type      = rt
  path      = /var/lib/sphinxsearch/data/irc_msg
  rt_attr_timestamp = timestamp
  rt_field    = nick
  rt_field    = channel
  rt_attr_uint    = channel_id
  rt_field    = msg
  rt_attr_uint    = line_index
  rt_attr_uint    = nick_id
  rt_attr_uint    = permission
  rt_attr_uint    = user_id
}
```

if you find yourself at a high line count and running low on ram, you could add ondisk_attrs = 1. this will remove all the uints above from ram, which means querying on just those will be slow. you should try to primarily query by field

## znc
running at commit 15ccaca41a17a06dfb5957156c03056524a71ae6
the regex would need updating if you changed the default log line format.
log path hasn't changed in a while but this code expects /home/user/.znc/users/a_user/networks/config_network/#channel/YYYY-MM-DD.log
