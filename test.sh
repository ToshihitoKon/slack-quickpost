#!/usr/bin/env bash
set -xeu

binpath="/tmp/slack-quickpost-bin"
go build -o $binpath ./...

testProfile=test-profile
slackToken=$(yq '.token' ~/.config/slack-quickpost/${testProfile}.yaml)
channelID=$(yq '.channel' ~/.config/slack-quickpost/${testProfile}.yaml)
iconUrl="https://user-images.githubusercontent.com/10419053/132874304-3c6397b5-e084-4476-a376-e3c7a941039c.png"

# print version
$binpath --version

unset SLACK_TOKEN
# token and channel given by Profile
$binpath --profile ${testProfile} \
    --text "case: token and channel given profile" \

export SLACK_QUICKPOST_PROFILE=${testProfile}
$binpath \
    --text "case: token and channel given profile 2" \
unset SLACK_QUICKPOST_PROFILE

# text / token given as option
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

cat << EOS > /tmp/slack-quickpost-test1.txt
case given text filepath.
EOS
$binpath --channel $channelID \
    --textfile /tmp/slack-quickpost-test1.txt

cat << EOS > /tmp/slack-quickpost-test2.txt
case given texitfile path as file.
post file mode, ignore username and icon.
EOS
$binpath --channel $channelID \
    --username "ignored username" \
    --icon "cry" \
    --file /tmp/slack-quickpost-test2.txt

wget -O /tmp/slack-quickpost-test.jpg $iconUrl
$binpath --channel $channelID  \
    --username "ignored username" \
    --icon "cry" \
    --file /tmp/slack-quickpost-test.jpg
