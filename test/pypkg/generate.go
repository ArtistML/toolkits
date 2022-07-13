package echo

// this is test example. image and registry is invalid.
//go:generate go run "github.com/artistml/toolkits/cmd/gen-pypkg" -p /tmp/toolkits/toolkits-python -i github.com/artistml/base-python:3.8.13 -r github.com/artistml --nexus-url=http://github.com/artistml --nexus-username=artistml --nexus-password=artistml-pw --nexus-pypi-path=pypi-hosted --pdm-source-name=pypi --pdm-source-url=http://github.com/artistml/repository/pypi-group/simple --pdm-author-name=artistml --pdm-author-email=artistml@github.com --pdm-license=ArtistML
