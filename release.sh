 #!/bin/bash
set -e

if [ "$1" = "-h" ]; then
    echo "use | $ ./release prod |  for a production release or just | $ ./release | for a snapshot release"
    exit 0
fi

if brew ls --versions goreleaser > /dev/null; then
  # The package is installed
  echo "."
else
    brew install goreleaser
fi

 if [ "$1" = "prod" ]; then
    echo "prodction release"
    goreleaser --rm-dist
 else
    echo "snapshot release"
    goreleaser --rm-dist --snapshot
 fi

