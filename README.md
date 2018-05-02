# znc-log-search
golang + arangodb + sphinx search = znc log search engine

this log search engine was originally written in python with rabbitmq and 3 python instances for 'performance'.
it was a total mess, and this isn't much better.

sphinxsearch was chosen due to ram constraints

to read any data for display, the original log files will need to be retained and not modified, the search index only retains file seek indexes.

config object description is in conf.go, config.json is your custom json file

## go

golang.org has a download link

## arangodb

create a server, user, password and db for this application. configure them in the config. create collections Nicks, Channels and Users. the application will blow up until they're accessible.

## sphinx config

index irc_msg
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
