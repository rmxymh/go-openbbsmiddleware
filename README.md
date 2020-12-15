# go-openbbsmiddleware
go implementation of [openbbs-middleware](https://hackmd.io/@twbbs/Root#%E6%9E%B6%E6%A7%8B%E5%9C%96).

這裡是使用 golang 來達成 [openbbs-middleware](https://hackmd.io/@twbbs/Root#%E6%9E%B6%E6%A7%8B%E5%9C%96).

## Getting Started

You can start with the [swagger api](http://173.255.216.176:5000)
and try the api.

You can copy the curl command from the link if you encounter
CORS issue.

你可以到 [swagger api](http://173.255.216.176:5000/)
並且試著使用 api.

如果你在 swagger 網頁裡遇到 CORS 的問題. 你可以在網頁裡 copy
curl 指令測試.

## Docker-compose

You can do the following to start with docker-compose:

* copy `docker_compose.env.template` to `docker_compose.env` and modify the settings.
* `./scripts/docker_initbbs.sh [BBSHOME] pttofficialapps/go-pttbbs:latest`
* `docker-compose --env-file docker_compose.env -f docker-compose.yaml up -d`
* register at `http://localhost:3457/account/register`
* login at `http://localhost:3457/account/login`
* `telnet localhost 8888` and use the account that you registered.

你可以使用以下方式來使用 docker-compose:

* 將 `./docker_compose.env.template` copy 到 `./docker_compose.env` 並且更改 BBSHOME 到你所希望的位置.
* `./scripts/docker_initbbs.sh [BBSHOME] pttofficialapps/go-pttbbs:latest`
* `docker-compose --env-file docker_compose.env -f docker-compose.yaml up -d`
* 在 `http://localhost:3457/account/register` 做 register
* 在 `http://localhost:3457/account/login` 做 login
* `telnet localhost 8888` 並且使用你剛剛登錄的帳號使用.

## Discussing / Reviewing / Questioning the code.

Besides creating issues, you can do the following
to discuss / review / question the code:

* `git clone` the repo
* create a review-[topic] branch
* commenting at the specific code-region.
* pull-request
* start discussion.
* close the pr with comments with only the link of the pr in the code-base.

除了開 issues 以外, 你還可以做以下的事情來對於 code 做討論 / review / 提出問題.

* `git clone` 這個 repo.
* 開一個 review-[topic] 的 branch.
* 對於想要討論的部分在 code 裡寫 comments.
* pull-request
* 對於 PR 進行討論.
* 當 PR 關掉時, comments 會留下關於這個 pr 討論的 link.

## Develop

You can start developing by `git clone` this repository.

你可以使用 `git clone` 來一起開發.

## Unit-Test

You can do unit-test with:

你可以做以下的事情來進行 unit-test:

* `./scripts/test.sh`

You can check coverage with:

你可以做以下的事情來進行 coverage-check:

* `./scripts/coverage.sh`


## Swagger

You can run swagger with:

你可以做以下的事情將 swagger 跑起來:

* `./scripts/swagger.sh`
* go to `http://localhost:5000`
