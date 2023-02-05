#!/bin/bash -e
set -f # avoid globbing because we have * on the rules lines
declare -A rules
declare -A output

OVERRIDE=${OVERRIDE:-false}
DEBUG=${DEBUG:-false}
TARGETS=$(bash -c "find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n'")

function build_db() {
  SCRIPT_PATH=$(dirname "$0")
  input="$SCRIPT_PATH/rules.txt"
  while IFS= read -r line
  do
    if [[ "$line" =~ ^\# ]] || [[ "$line" == "" ]]; then
      continue
    fi

    array=($line)
    if [[ ${#array[@]} != 3 ]] || [[ ! ${array[1]} == "-" ]]; then
      echo "ERROR: line format must be 'directory - value', line: $line"
      exit 1
    fi

    dir=${array[0]}
    value=${array[2]}

    case $value in
      all)
        rules[$dir]="${TARGETS[@]}"
      ;;
      none)
        rules[$dir]="none"
      ;;
      !*) # all beside x
        value="${value:1}" # remove first char
        rules[$dir]="${TARGETS[@]/$value}"
      ;;
      regex)
        regex_targets=($(bash -c "ls -d $dir | xargs -n1 basename | tr '\n' ' '"))
        # remove last dir x/y -> x/
        dir=$(echo $dir | sed 's![^/]*$!!')
        for element in "${regex_targets[@]}"; do
          rules[${dir}${element}/*]="${element/#k8s-}"
        done
      ;;
      regex_none)
        regex_targets=($(bash -c "ls -d $dir | tr '\n' ' '"))
        for element in "${regex_targets[@]}"; do
          rules[${element}/*]="none"
        done
      ;;
      *) # value as is
        rules[$dir]="$value"
      ;;
    esac
  done < "$input"
}

function printdb {
  printf '%.0s-' {1..50} && echo
  for str in "${!rules[@]}"; do
      echo "$str - $(echo ${rules[$str]})"
  done
  printf '%.0s-' {1..50} && echo
}

function matcher() {
  str=$1

  # perfect match of the file full path
  if [ "${rules[$str]}" != "" ]; then
    echo "${rules[$str]}"
    return
  fi

  # match of the dirname x/y/z.txt -> x/y
  if [ "${rules[$(dirname $str)]}" != "" ]; then
    echo "${rules[$(dirname $str)]}"
    return
  fi

  while [[ "$str" != "." ]]; do
    str=$(dirname $str)
    key=$str/*
    if [ "${rules[$key]}" != "" ]; then
      echo "${rules[$key]}"
      return
    fi
  done
}

function process() {
  filter=$1

  if ! git show $PREV_KUBEVIRTCI_TAG > /dev/null 2>&1; then
    echo "ERROR: $PREV_KUBEVIRTCI_TAG is missing, run 'git pull upstream main --tags'"
    exit 1
  fi

  list=($(git diff --name-only --diff-filter=$filter $PREV_KUBEVIRTCI_TAG | grep -vE "\.md$" || true))
  for file in "${list[@]}"; do
    targets="$(matcher "$file")"
    if [ "$targets" == "" ] && [ $filter == $ALL_BESIDE_DELETED ]; then
      echo "ERROR: match not found for $file (possible rules.txt update needed)"
      error_found=1
      continue
    fi

    if [ "$targets" == "" ]; then
      targets="none"
    fi

    [ $DEBUG == true ] && echo INFO: $file - $(echo "${targets[@]}")

    if [ "$targets" == "none" ]; then
      continue
    fi

    for target in ${targets[@]}; do
      if [ "${output[$target]}" != "" ]; then
        continue
      fi

      found=0
      for official_target in ${TARGETS[@]}; do
          if [ "$official_target" == "$target" ]; then
              found=1
              break
          fi
      done
      if [ $found == 0 ]; then
        echo "ERROR: target not valid: $target, file: $file (hint: leftovers of an old provider?)"
        error_found=1
        continue
      fi

      output[$target]=$target
    done
  done
}

function diff_finder() {
  TAG_URL="https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest?ignoreCache=1"
  PREV_KUBEVIRTCI_TAG=${PREV_KUBEVIRTCI_TAG:-$(curl -sL $TAG_URL)}

  ALL_BESIDE_DELETED="d" # all files beside deleted files, must have a target for them
  DELETED_ONLY="D" # deleted files only, allowed to not have a target for them

  error_found=0
  process $ALL_BESIDE_DELETED
  process $DELETED_ONLY

  if [ $error_found == 1 ]; then
    echo "ERROR: errors were found, exiting"
    exit 1
  fi

  [ $DEBUG == true ] && printf '%.0s-' {1..50} && echo
  if [ "${#output[@]}" -ne 0 ]; then
    echo ${!output[*]}
  else
    echo none
  fi
}

function check_override() {
  if [ "$OVERRIDE" == "all" ]; then
    echo $(echo "${TARGETS[@]}")
    exit 0
  fi

  if [ "$OVERRIDE" != false ]; then
    echo "$OVERRIDE"
    exit 0
  fi
}

function main() {
  check_override
  build_db
  [ $DEBUG == true ] && printdb
  diff_finder
}

main "$@"
