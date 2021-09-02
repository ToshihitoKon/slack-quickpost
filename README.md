# slack-quickpost

## installation

```
go get github.com/ToshihitoKon/slack-quickpost
```

## setup

https://api.slack.com/apps/  

make Slack App and get OAuth token.

## usase

OAuth token set environment variable or command option.

```
export SLACK_TOKEN="xoxb-XXXXXXXX-XXXXXXX-XXXXXX"
slack-quickpost \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

OR

```
slack-quickpost \
  --token xoxb-XXXXXXXX-XXXXXXX-XXXXXX \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

## comamnd options


```
--token OAuth token: require if not set SLACK_TOKEN
--channel CHANNEL_ID: require
--text TEXT: require
--username USERNAME: optional
--icon EMOJI_NAME: optional choose either --icon-url
--icon-url IMAGE_URL: optional choose either --icon
```
