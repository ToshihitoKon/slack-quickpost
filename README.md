# slack-quickpost

## installation

```
go install github.com/ToshihitoKon/slack-quickpost@latest
```

## setup

https://api.slack.com/apps/  

make Slack App and get OAuth token.

### GitHub Actions

```yaml
jobs:
  using-slack-quikpost:
    runs-on: ubuntu-latest
    steps:
      - uses: ToshihitoKon/slack-quickpost@v1
        with:
          version: 0.7.0

```

## usage

OAuth token given by one of the following methods.

#### environment variable

```
export SLACK_TOKEN="xoxb-XXXXXXXX-XXXXXXX-XXXXXX"
slack-quickpost \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

#### Option

```
slack-quickpost \
  --token xoxb-XXXXXXXX-XXXXXXX-XXXXXX \
  --channel [CHANNEL_ID] \
  --text [TEXT]
```

#### Config file

Save config yaml `~/.config/slack-quickpost/profile-name.yaml`

```yaml
token: xoxb-XXX
channel: XXX
```

Provide the profile name using an environment variable(`SLACK_QUICKPOST_PROFILE`) or option.

```
SLACK_QUICKPOST_PROFILE=profile-name slack-quickpost --text [TEXT]
slack-quickpost --profile profile-name --text [TEXT]
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

# post BlockKit
slack-quickpost \
  --channel [CANNEL_ID] \
  --block '{"type":"section","text":{"type":"mrkdwn","text":"*Sample BlockKit"}}'
```

### post file

```
slack-quickpost \
  --channel [CANNEL_ID] \
  --file [FILE_PATH]
```

## comamnd options

```
--blocks string      post BlockKit json
--channel string     post slack channel id
--file string        post file path
--icon string        icon emoji
--icon-url string    icon image url
--nofail             always return success code(0)
--profile string     slack quickpost profile name
--snippet            post text as snippet
--text string        post text
--textfile string    post text file path
--thread-ts string   post under thread
--token string       slack app OAuth token
--username string    user name
--version            print version
```
