# Toggl test task

An application that fetches each test taker and sends an email 
to the non-demo test takers that scored at least 80%.
The service wakes up every 10 minutes (depending on the configuration).

## Configuration

An example configuration is provided below:
```shell script
SENDER_EMAIL=miklos@toggl.com
SENDER_PASSWORD=123456
SENDER_TAKERS_API=http://th-hw.herokuapp.com/api/v1
SENDER_REDIS_ADDR=localhost:6379
SENDER_SERVER_UPDATE_INTERVAL=10m     # interval between reloading takers
SENDER_SERVER_POLL_INTERVAL=3m        # interval betwen redis polls
SENDER_SERVER_SEND_THANKS_TIMEOUT=20s # timeout for each message sending operation (redis + email)
SENDER_LOG_LEVEL=info                 # see logrus.Level
```