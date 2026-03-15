VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o dotfiles .

test:
	go test -v ./...

release: clean test
	@mkdir -p dist
	cp dotfiles.1 dist/
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/dotfiles . && \
		tar -czf dist/dotfiles-$(VERSION)-linux-amd64.tar.gz -C dist dotfiles dotfiles.1 && rm dist/dotfiles
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/dotfiles . && \
		tar -czf dist/dotfiles-$(VERSION)-linux-arm64.tar.gz -C dist dotfiles dotfiles.1 && rm dist/dotfiles
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/dotfiles . && \
		tar -czf dist/dotfiles-$(VERSION)-darwin-amd64.tar.gz -C dist dotfiles dotfiles.1 && rm dist/dotfiles
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/dotfiles . && \
		tar -czf dist/dotfiles-$(VERSION)-darwin-arm64.tar.gz -C dist dotfiles dotfiles.1 && rm dist/dotfiles
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/dotfiles.exe . && \
		cd dist && zip dotfiles-$(VERSION)-windows-amd64.zip dotfiles.exe dotfiles.1 && rm dotfiles.exe
	rm dist/dotfiles.1

clean:
	rm -rf dist/
	rm -f dotfiles

deploy: build install-man install-completion
	cp dotfiles ~/.local/bin/

install-man:
	install -d /usr/local/share/man/man1
	install -m 644 dotfiles.1 /usr/local/share/man/man1/dotfiles.1

install-completion:
	install -d ~/.oh-my-zsh/custom/completions
	install -m 644 _dotfiles ~/.oh-my-zsh/custom/completions/_dotfiles

.PHONY: build test release clean deploy install-man install-completion
