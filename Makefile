.PHONY: gofmt
## run gofmt format code
gofmt:
	@printf "🐋 \033[1;32m===> Go format ...\033[0m\n"
	gofmt -w -s ./