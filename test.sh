#!/usr/bin/env bash
set -euo pipefail

tempdir=$(mktemp -d)
trap 'rm -rf "$tempdir"' EXIT

binpath="${tempdir}/slack-quickpost-bin"
go build -o "$binpath" .

testProfile=${SLACK_QP_PROFILE:-"test-profile"}
profilePath="${HOME}/.config/slack-quickpost/${testProfile}.yaml"
slackToken=$(yq -r '.token' "$profilePath")
channelID=$(yq -r '.channel' "$profilePath")
iconUrl="https://user-images.githubusercontent.com/10419053/132874304-3c6397b5-e084-4476-a376-e3c7a941039c.png"

post() {
    "$binpath" --channel "$channelID" "$@"
}

# print version
"$binpath" --version

unset SLACK_TOKEN
# token and channel given by Profile
"$binpath" --profile "$testProfile" \
    --text "case: token and channel given profile"

SLACK_QUICKPOST_PROFILE=$testProfile "$binpath" \
    --text "case: token and channel given profile 2"

"$binpath" --profile "${testProfile}.yaml" \
    --text "case: profile name given with .yaml extension"

# text / token given as option
post --text "case: token given command option, username and icon not given" \
    --token "$slackToken"

export SLACK_TOKEN=$slackToken

post --text "case: token given environment variable, username given" \
    --username "given username 1"

post --text "given icon with emoji" \
    --username "given username_2" \
    --icon "thumbsup"

post --text "given icon with image url" \
    --username "given username 3" \
    --icon-url "$iconUrl"

post --text "case post text with snippet option\nsnippet name become timestamp" \
    --username "given username 4" \
    --icon "tada" \
    --snippet

cat << EOS > "${tempdir}/slack-quickpost-test1.txt"
case given text filepath.
EOS
post --textfile "${tempdir}/slack-quickpost-test1.txt"

cat << EOS > "${tempdir}/slack-quickpost-test2.txt"
case given textfile path as file.
post file mode, ignore username and icon.
EOS
post --username "ignored username" \
    --icon "cry" \
    --file "${tempdir}/slack-quickpost-test2.txt"

curl -fsSL -o "${tempdir}/slack-quickpost-test.jpg" "$iconUrl"
post --username "ignored username" \
    --icon "cry" \
    --file "${tempdir}/slack-quickpost-test.jpg"

# BlockKit json
post --username "BlockKit json" \
    --icon "blush" \
    --blocks '[{"type":"section","text":{"type":"mrkdwn","text":"*Post BlockKit*"}}]'

# delete message by permalink
# permalink のホスト部は削除対象の特定に使われないため、slack.com で組み立てる
postOutput=$(post --text "case: this message will be deleted")
deleteChannel=$(echo "$postOutput" | yq -r '.channel')
deleteTs=$(echo "$postOutput" | yq -r '.timestamp')
deleteOutput=$("$binpath" --delete "https://slack.com/archives/${deleteChannel}/p${deleteTs//./}")
[ "$deleteOutput" = "$postOutput" ]

# delete snippet / file by file link
# ファイル URL のユーザー部とファイル名も削除対象の特定に使われないため、ダミーで組み立てる
deleteUploaded() {
    local fileID
    fileID=$(echo "$1" | yq -r '.file')
    deleteOutput=$("$binpath" --delete "https://slack.com/files/U0/${fileID}/deleted")
    [ "$deleteOutput" = "{\"file\":\"${fileID}\"}" ]
}

snippetOutput=$(post --text "case: this snippet will be deleted" --snippet)
deleteUploaded "$snippetOutput"

cat << EOS > "${tempdir}/slack-quickpost-test3.txt"
case: this file will be deleted.
EOS
fileOutput=$(post --file "${tempdir}/slack-quickpost-test3.txt")
deleteUploaded "$fileOutput"
