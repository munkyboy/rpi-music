.DEFAULT_GOAL := target/rpi.img
.PHONY: clean cache_clean run copy

target/librespot: Dockerfile-librespot
	@-mkdir -p target
	docker build -t librespot-armv7 -f Dockerfile-librespot .
	@docker rm -f librespot-armv7 &> /dev/null
	docker create --name librespot-armv7 librespot-armv7
	docker cp librespot-armv7:/opt/src/librespot/target/armv7-unknown-linux-gnueabihf/release/librespot ./target/librespot
	@touch target/librespot
	@docker rm -f librespot-armv7 &> /dev/null

amp-httpd/target/amp-httpd-rpi:
	cd amp-httpd && make target/amp-httpd-rpi

SSH_PUB_KEY ?= $(HOME)/.ssh/id_rsa.pub
target/rpi.img: rpi.pkr.hcl scripts/* target/librespot *.auto.pkrvars.hcl amp-httpd/target/amp-httpd-rpi
	@mkdir -p target
	docker run --rm \
		--privileged -v /dev:/dev \
		-v rpi-music-packer-cache:/tmp/packer_cache \
		-e PACKER_CACHE_DIR=/tmp/packer_cache \
		-v $(SSH_PUB_KEY):/root/.ssh/id_rsa.pub \
		-v $(PWD):/build --workdir /build \
		mkaczanowski/packer-builder-arm \
		build .

clean:
	rm -rf target

cache_clean:
	docker volume rm rpi-music-packer-cache

copy: target/rpi.img
	@diskutil list /dev/disk2
	@echo "#########\ncopying to /dev/disk2\n#########"
	@read -p "Does this look correct? (y/n) " INPUT; if [ "$$INPUT" != "y" ] ; then echo "aborting"; exit 1 ; fi
	@echo "continuing"
	@diskutil unmountDisk /dev/disk2
	sudo dd bs=1m if=target/rpi.img of=/dev/disk2

run: target/rpi.img
	docker run -it -v $(PWD)/target/rpi.img:/sdcard/filesystem.img lukechilds/dockerpi:vm pi3
