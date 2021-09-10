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

### post text

```
slack-quickpost \
  --channel [CANNEL_ID] \
  --text [TEXT] \
  --username [DISPLAY_USERNAME] \
  --icon [ICON_EMOJI] 

# text given textfile path and icon given image url
slack-quickpost \
  --channel [CANNEL_ID] \
  --textfile [FIlEPATH] \
  --username [DISPLAY_USERNAME] \
  --icon_url [ICON_IMAGE_URL] 

# post text as snippet
slack-quickpost \
  --channel [CANNEL_ID] \
  --text [TEXT] \
  --snippet
```

### post file

```
slack-quickpost \
  --channel [CANNEL_ID] \
  --file [FILE_PATH]
```

## comamnd options

```
--channel string    post slack channel id
--file string       post file path
--icon string       icon emoji
--icon-url string   icon image url
--snippet           post text as snippet
--text string       post text
--textfile string   post text file path
--token string      slack app OAuth token
--username string   user name
--version           print version
```
