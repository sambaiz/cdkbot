if [ $1 = "list" ]; then
  echo -e "Stack1\nStack2"
elif [ $1 = "diff" ]; then
  echo -e "diff\ndiff"
  exit 1
elif [ $1 = "deploy" ]; then
  echo -e "deploy\ndeploy"
fi