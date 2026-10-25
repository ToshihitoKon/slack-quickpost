#compdef slack-quickpost
compdef _slack-quickpost slack-quickpost

_slack-quickpost_profiles() {
  local expl
  local -a profiles
  profiles=("$HOME"/.config/slack-quickpost/*.yaml(N:t:r))
  _wanted profiles expl 'profile' compadd -a profiles
}

_slack-quickpost() {
  _arguments \
    '(- *)'{-h,--help}'[show help]' \
    '(- *)--version[print version]' \
    '(- *)--completion=[print shell completion script (bash or zsh)]:shell:(bash zsh)' \
    '--blocks=[post BlockKit json]:json: ' \
    '--channel=[post slack channel id]:channel id: ' \
    '--delete=[delete message by permalink url]:url: ' \
    '--file=[post file path]:file:_files' \
    '--icon=[icon emoji]:emoji: ' \
    '--icon-url=[icon image url]:url: ' \
    '--nofail[exit with code 0 even on failure (except flag parse errors)]' \
    '--profile=[slack quickpost profile name]:profile:_slack-quickpost_profiles' \
    '--snippet[post text as snippet]' \
    '--text=[post text]:text: ' \
    '--textfile=[post text file path]:file:_files' \
    '--thread-ts=[post under thread]:thread ts: ' \
    '--token=[slack app OAuth token]:token: ' \
    '--username=[user name]:user name: '
}

# source / eval で読み込んだときに補完の外で _arguments を実行しないよう、fpath から autoload されたときだけ呼ぶ
if [ "$funcstack[1]" = "_slack-quickpost" ]; then
  _slack-quickpost "$@"
fi
