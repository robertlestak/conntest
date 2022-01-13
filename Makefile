bin: bin/conntest_darwin bin/conntest_windows bin/conntest_linux

bin/conntest_darwin:
	GOOS=darwin go build -o $@
bin/conntest_windows:
	GOOS=windows go build -o $@
bin/conntest_linux:
	GOOS=linux go build -o $@

docker:
	docker build -f devops/docker/Dockerfile -t docker-registry.umusic.com/devops/conntest:latest .

.PHONY: docker