.DEFAULT_GOAL := target/rpi.img
.PHONY: clean cache_clean run copy

target/librespot: Dockerfile-librespot
	@-mkdir -p target
	docker build --load -t librespot-armv7 -f Dockerfile-librespot .
	@-docker rm -f librespot-armv7 &> /dev/null
	docker create --name librespot-armv7 librespot-armv7
	docker cp librespot-armv7:/opt/src/librespot/target/armv7-unknown-linux-gnueabihf/release/librespot ./target/librespot
	@touch target/librespot
	@docker rm -f librespot-armv7 &> /dev/null

target/shairport-sync: Dockerfile-shairport
	@-mkdir -p target
	docker build --load --platform linux/arm/v7 -t shairport-armv7 -f Dockerfile-shairport .
	@-docker rm -f shairport-armv7 &> /dev/null
	docker create --platform linux/arm/v7 --name shairport-armv7 shairport-armv7
	docker cp shairport-armv7:/usr/local/bin/nqptp target/
	docker cp shairport-armv7:/usr/local/bin/shairport-sync target/
	@touch target/nqptp target/shairport-sync
	@docker rm -f shairport-armv7 &> /dev/null

amp-httpd/target/amp-httpd-rpi:
	cd amp-httpd && make target/amp-httpd-rpi

SSH_PUB_KEY ?= $(HOME)/.ssh/id_rsa.pub
target/rpi.img: rpi.pkr.hcl scripts/* target/librespot *.auto.pkrvars.hcl amp-httpd/target/amp-httpd-rpi target/shairport-sync target/nqptp
	@mkdir -p target
	docker run --rm -it \
		--privileged -v /dev:/dev \
		-v rpi-music-packer-cache:/tmp/packer_cache \
		-e PACKER_CACHE_DIR=/tmp/packer_cache \
		-e PACKER_LOG=1 \
		-v $(SSH_PUB_KEY):/root/.ssh/id_rsa.pub \
		-v $(PWD):/build --workdir /build \
		mkaczanowski/packer-builder-arm@sha256:b04b5a0b3d03b12b57ffa8f89bd5fed76084f16864c226e0426f773b43dbcafa \
		build -on-error=ask .

clean:
	rm -rf target

cache_clean:
	docker volume rm rpi-music-packer-cache

copy: target/rpi.img
	# TODO: make this non-macos specific
	@diskutil list /dev/disk2
	@echo "#########\ncopying to /dev/disk2\n#########"
	@read -p "Does this look correct? (y/n) " INPUT; if [ "$$INPUT" != "y" ] ; then echo "aborting"; exit 1 ; fi
	@echo "continuing"
	@diskutil unmountDisk /dev/disk2
	sudo dd bs=1m if=target/rpi.img of=/dev/disk2

run: target/rpi.img
	@echo "to connect via SSH, run 'ssh pi@localhost -p 5022'"
	docker run -it -v $(PWD)/target/rpi.img:/sdcard/filesystem.img -p 5022:5022 lukechilds/dockerpi:vm pi3
