# slack-quickpost

## setup

https://api.slack.com/apps/
make Slack App and set OAuth token

```
export SLACK_TOKEN="xoxb-XXXXXXXX-XXXXXXX-XXXXXX"
```

## usase

```
slack-quickpost \
  --channel [CHANNEL_ID:require] \
  --text [TEXT:require] \
  --username [USERNAME:optional] \
  --icon [EMOJI_NAME:optional]
```
