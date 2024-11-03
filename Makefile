export WAIT_DELAY=20s
export DISCORD_WEBHOOK_USERNAME=Factorio

.PHONY: run
run: build
	./factorio-notify-bot2 _bin/factorio/bin/x64/factorio --start-server _bin/map.zip

.PHONY: setup
setup:
	rm -rf _bin
	mkdir -p _bin
	cd _bin ; { \
		wget -O factorio.tar.xz https://factorio.com/get-download/stable/headless/linux64 ; \
		tar xf factorio.tar.xz ; \
		factorio/bin/x64/factorio --create map ; \
	}

.PHONY: build
build:
	go build
