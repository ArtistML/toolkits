.PHONY: lint
## run lint check
lint:
	@printf "🔎 \033[1;32m===> Running go lint for all packages...\033[0m\n"
	@printf "\033[33mNote: lint is also included in \`review\`, this is just a convenient target for local run.\033[0m\n"
	golangci-lint run
	
.PHONY: gofmt
## run gofmt format code
gofmt:
	@printf "🐋 \033[1;32m===> Go format ...\033[0m\n"
	gofmt -w -s ./
	goimports -local github.com/artistml -w ./

