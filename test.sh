#!/usr/bin/env bash
set -xeu

binpath="/tmp/slack-quickpost-bin"
go build -o $binpath ./...

slackToken=$SLACK_QUICKPOST_TEST_TOKEN
channelID=$SLACK_QUICKPOST_TEST_CHANNEL_ID
iconUrl="https://user-images.githubusercontent.com/10419053/132874304-3c6397b5-e084-4476-a376-e3c7a941039c.png"

# print version
$binpath --version

# text / token given as option
unset SLACK_TOKEN
$binpath --channel $channelID \
    --text "case: token given command option, username and icon not given" \
    --token="${slackToken}"

export SLACK_TOKEN=$slackToken

$binpath --channel $channelID \
    --text "case: token given environment variable, username given" \
    --username "given username 1"

$binpath --channel $channelID \
    --text "given icon with emoji" \
    --username "given username_2" \
    --icon "thumbsup"

$binpath --channel $channelID \
    --text "given icon with image url" \
    --username "given username 3" \
    --icon-url $iconUrl

$binpath --channel $channelID \
    --text "case post text with snippet option\nsnippet name become timestamp" \
    --username "given username 4" \
    --icon "tada" \
    --snippet

cat << EOS > /tmp/slack-quickpost-test.txt
case given text filepath.
EOS
$binpath --channel $channelID \
    --textfile /tmp/slack-quickpost-test

cat << EOS > /tmp/slack-quickpost-test.txt
case given texitfile path as file.
post file mode, ignore username and icon.
EOS
$binpath --channel $channelID \
    --username "ignored username" \
    --icon "cry" \
    --text "text file" \
    --file /tmp/slack-quickpost-test

wget -O /tmp/slack-quickpost-test.jpg $iconUrl
$binpath --channel $channelID  \
    --username "ignored username" \
    --icon "cry" \
    --text "image file" \
    --file /tmp/slack-quickpost-test.jpg
