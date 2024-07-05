.DEFAULT_GOAL := target/rpi.img
.PHONY: clean cache_clean run copy

target/librespot: Dockerfile-librespot
	@-mkdir -p target
	docker build --load -t librespot-arm64 -f Dockerfile-librespot .
	@-docker rm -f librespot-arm64 &> /dev/null
	docker create --name librespot-arm64 librespot-arm64
	docker cp librespot-arm64:/opt/src/librespot/target/aarch64-unknown-linux-gnu/release/librespot ./target/librespot
	@touch target/librespot
	@docker rm -f librespot-arm64 &> /dev/null

target/shairport-sync: Dockerfile-shairport
	@-mkdir -p target
	docker build --load --platform linux/arm64 -t shairport-arm64 -f Dockerfile-shairport .
	@-docker rm -f shairport-arm64 &> /dev/null
	docker create --platform linux/arm64 --name shairport-arm64 shairport-arm64
	docker cp shairport-arm64:/usr/local/bin/nqptp target/
	docker cp shairport-arm64:/usr/local/bin/shairport-sync target/
	@touch target/nqptp target/shairport-sync
	@docker rm -f shairport-arm64 &> /dev/null

amp-httpd/target/amp-httpd-rpi:
	cd amp-httpd && make target/amp-httpd-rpi

SSH_PUB_KEY ?= $(HOME)/.ssh/id_rsa.pub
target/rpi.img: rpi.pkr.hcl $(wildcard scripts/*) target/librespot $(wildcard *.auto.pkrvars.hcl) amp-httpd/target/amp-httpd-rpi target/shairport-sync target/nqptp
	@mkdir -p target
	docker run --rm -it \
		--privileged -v /dev:/dev \
		-v rpi-music-packer-cache:/tmp/packer_cache \
		-e PACKER_CACHE_DIR=/tmp/packer_cache \
		-e PACKER_LOG=1 \
		-v $(SSH_PUB_KEY):/root/.ssh/id_rsa.pub \
		-v $(PWD):/build --workdir /build \
		mkaczanowski/packer-builder-arm:1.0.9 \
		build -on-error=ask .

clean:
	rm -rf target

cache_clean:
	docker volume rm rpi-music-packer-cache

TARGET_DEVICE=/dev/disk4
copy: target/rpi.img
	# TODO: make this non-macos specific
	@diskutil list $(TARGET_DEVICE)
	@echo "#########\ncopying to $(TARGET_DEVICE)\n#########"
	@read -p "Does this look correct? (y/n) " INPUT; if [ "$$INPUT" != "y" ] ; then echo "aborting"; exit 1 ; fi
	@echo "continuing"
	@diskutil unmountDisk $(TARGET_DEVICE)
	sudo dd bs=1m if=target/rpi.img of=$(TARGET_DEVICE)

# Support for running image via QEMU
## convert img to qcow
target/rpi.qcow2: target/rpi.img
	qemu-img convert -f raw -O qcow2 $< $@
	qemu-img resize $@ 4G

## files embedded in raspbian image extracted for
device_tree_blob = bcm2710-rpi-3-b-plus.dtb
kernel_img = kernel8.img
target/$(kernel_img): target/rpi.img
	docker run -it --rm --privileged -v $$PWD/target:/opt/target -w /opt alpine ash -c "mkdir -p boot; mount -o loop,offset=4194304 target/rpi.img boot; cp boot/$(device_tree_blob) boot/$(kernel_img) target/"

run: target/$(kernel_img) target/rpi.qcow2
	@echo "after this is up, run: ssh pi@localhost -p 2222"
	qemu-system-aarch64 \
		-machine raspi3b \
		-serial stdio \
		-dtb target/$(device_tree_blob) \
		-kernel target/$(kernel_img) \
		-sd target/rpi.qcow2 \
		-append "earlyprintk loglevel=8 console=ttyAMA0,115200 dwc_otg.lpm_enable=0 root=/dev/mmcblk0p2 rw rootdelay=1" \
		-device usb-net,netdev=net0 -netdev user,id=net0,hostfwd=tcp::2222-:22 \
		-usb -device usb-mouse -device usb-kbd