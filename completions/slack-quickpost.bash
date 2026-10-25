_slack_quickpost() {
  local cur=${COMP_WORDS[COMP_CWORD]}
  local prev=${COMP_WORDS[COMP_CWORD-1]}

  # COMP_WORDBREAKS の = で --profile=wo が --profile, =, wo に分かれ、prev が = になるため、= の前の語を見る
  if [[ $cur == = ]]; then
    cur=
  elif [[ $prev == = ]]; then
    prev=${COMP_WORDS[COMP_CWORD-2]}
  fi

  case $prev in
    --profile)
      # compgen -W は候補の文字列を展開するため、ファイル名を渡さずグロブで絞り込む
      local f
      COMPREPLY=()
      for f in "$HOME"/.config/slack-quickpost/"$cur"*.yaml; do
        [ -e "$f" ] || continue
        f=${f##*/}
        COMPREPLY+=("${f%.yaml}")
      done
      return
      ;;
    --completion)
      COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
      return
      ;;
  esac

  if [[ $cur == -* ]]; then
    COMPREPLY=($(compgen -W "--blocks --channel --completion --delete --file --help --icon --icon-url --nofail --profile --snippet --text --textfile --thread-ts --token --username --version" -- "$cur"))
  fi
}

# compgen -f で返すとディレクトリに / が付かず、付けるための compopt -o filenames は macOS 標準の bash 3.2 にないため、
# --file などの値は COMPREPLY を空にして -o default のファイル名補完に任せる（--text など他の値でもファイル名が候補に出る）
complete -o default -F _slack_quickpost slack-quickpost
