# bard_agent_api_go

This is an implementation of [Bard's agent-api](https://github.com/bard-rr/agent-api) written in Go using the [Gin framework.](https://gin-gonic.com/)

Just for fun, this branch will have a redis implementation. The goal will be to improve the API's performance: rather than hit the db to decide if a session is new or not, we'll just check a cache for ended sessions and a cache for active sessions.

## Redis Reference

`sudo service redis-server start` start redis
`redis-cli ping`: check to see if the server is up.
