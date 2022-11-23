# bard_agent_api_go

This is an implementation of [Bard's agent-api](https://github.com/bard-rr/agent-api) written in Go using the [Gin framework.](https://gin-gonic.com/)

## Redis

Just for fun, this implementaiton also utilizes redis with the goal of improving the API's performance during a load test. Redis use looks like this:

- Two redis caches: one for active sessions, one for ended sessions
- The redis cache for active sessions will store the most recent event timestamp
- The redis cache for ended sessions will store nothing

This way, the agent-api won't need to hit Clickhouse at all, and it only needs to hit postgres while creating a new session. It can tell if a session exists by checking each cache, and it can update the most recent event time by editing the redis cache for active sessions. That should remove 99% of the database hits that were likely the heaviest performance drag in the agent-api's previous implementation.

## Future Directions (?)

If we wanted to integrate this version of the agent-api with the rest of bard, we'd need to make some changes. Could go in a few directions, but here are some immediate thoughts

- The session ender will need to populate the cache for ended sessions after it moves a session from postgres to clickhouse
- The session ender will also need to change how it decides if a session has ended: maybe it just checks the cache value for the most recent event time of each session, maybe part of what it does is update the most recent event time in postgres for each session?
- I've taken no steps to persist the data stored in the redis caches. That data is an integral part of this implementation, so it **CANNOT** be ephemeral.

### Redis Reference

Here are some helpful commands for working with redis

`sudo service redis-server start` start redis
`redis-cli ping`: check to see if the server is up.
`redis-cli` takes you to a cli tool with lots of helpful commands

- `get <key>` will give you the key
- `set <key> <value>` will set the given key to the given value
- `FLUSHALL ASYNC` will delete all keys
