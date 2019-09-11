if [ "$1" == "list" ]; then
  echo -e "Stack1\nStack2"
elif [ "$2" == "failStack" ]; then
  echo "failed!"
  exit 2
elif [ "$1" == "diff" ]; then
  echo -e "diff: $@"
  if [ "$2" == "diffStack" ]; then
    exit 1 # has diff
  fi
  exit 0 # has no diff
elif [ "$1" == "deploy" ]; then
  echo -e "deploy: $@"
fi