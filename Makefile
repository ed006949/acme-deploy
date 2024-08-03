DATE		=	`date`
GIT_STATUS	=	`git status --short`

all:	commit
all:	build
all:	status

build:
	go build -ldflags="-s -w" -trimpath -o "./bin/${NAME}" ./src/*.go
	GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o "./bin/${NAME}-freebsd-amd64" ./src/*.go

clean:
	-gh auth logout
	-go clean -i -r -x -cache -testcache -modcache -fuzzcache
	-rm -v go.mod
	-rm -v go.sum
	-find ./ -name ".DS_Store" -delete
	-find ./ -name "._.DS_Store" -delete

commit: status
#
# TODO
#ifneq (${GIT_STATUS},)
#
ifneq ($(shell git status --short),)
	git add .
	git commit -m "Makefile commit (${DATE})"
	git push
endif

diff:
	git diff

execute:
	./bin/${NAME} ${COMMAND_LINE}

init:
	go mod init ${TARGET}
	go get -u ./...
	go mod tidy

install:
	@echo ${NAME} ${PACKAGE} ${TARGET} ${DATE} ${GIT_STATUS}

race:
	go run -race ./... ${COMMAND_LINE}

release: commit
	git tag v${VERSION}
	git push origin v${VERSION}
	gh release create v${VERSION} --generate-notes --latest=true

run:
	go run -ldflags="-s -w" -trimpath ./... ${COMMAND_LINE}

status:
	git status

test:
	go test ./...

update:
	go get -u ./...
	go mod tidy

include Makefile.local

#
# possibly destructive actions
# possibly destructive actions
#

#
#
#
#
gitignore:
	curl -o ./.gitignore ${GITIGNORE_URL}
	cat ./.local.gitignore >> ./.gitignore

#
# init local package
# > make init_localpackage localpackage=package_name
#
init_localpackage:
ifneq (${localpackage},)
	mkdir ./src/${localpackage}
#	echo "package ${localpackage}" > ./src/${localpackage}/${localpackage}.go
	echo "package ${localpackage}" > ./src/${localpackage}/const.go
	echo "package ${localpackage}" > ./src/${localpackage}/errors.go
	echo "package ${localpackage}" > ./src/${localpackage}/func.go
	echo "package ${localpackage}" > ./src/${localpackage}/init.go
	echo "package ${localpackage}" > ./src/${localpackage}/method.go
	echo "package ${localpackage}" > ./src/${localpackage}/type.go
	echo "package ${localpackage}" > ./src/${localpackage}/var.go
endif

#
# new repo init
# > make init-init-init-init
#
init-init-init-init:	clean
	-gh auth logout
	gh auth login --with-token < ~/.git_token
	-gh repo delete ${NAME} --yes
	-rm -Rfv ./.git
	git init
	git config commit.gpgSign false
	gh repo create ${NAME} --private --source=.
#	git config --add --bool push.autoSetupRemote true
#ifneq (${GPG_KEY},)
#	git config --add --bool commit.gpgSign true
#	git config --add --string user.signingkey ${GPG_KEY}
#endif
	git add .
	git commit -m "Makefile initial commit (${DATE})"
	git push --set-upstream origin master
#	git push
	go mod init ${TARGET}
	go get -u ./...
	go mod tidy
	git add .
	git commit -m "Makefile initial update (${DATE})"
	git push
