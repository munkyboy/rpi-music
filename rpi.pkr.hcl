variable "wifi_ssid" {
  type = string
}

variable "wifi_password" {
  type      = string
  sensitive = true
}

# must match path in `/usr/share/zoneinfo`
variable "tz" {
  type    = string
  default = "America/Los_Angeles"
}

# Example: https://github.com/mkaczanowski/packer-builder-arm/blob/master/boards/raspberry-pi/raspbian-resize.json
source "arm" "rpi" {
  file_urls             = ["https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2022-04-07/2022-04-04-raspios-bullseye-armhf-lite.img.xz"]
  file_target_extension = "xz"
  file_unarchive_cmd    = ["xz", "-d", "$ARCHIVE_PATH"]
  file_checksum_type    = "sha256"
  file_checksum         = "34987327503fac1076e53f3584f95ca5f41a6a790943f1979262d58d62b04175"
  image_build_method    = "resize"
  image_path            = "/opt/build.img"
  image_size            = "3G"
  image_type            = "dos"
  image_partitions {
    name         = "boot"
    type         = "c"
    start_sector = "8192"
    filesystem   = "vfat"
    size         = "256M"
    mountpoint   = "/boot"
  }
  image_partitions {
    name         = "root"
    type         = "83"
    start_sector = "532480"
    filesystem   = "ext4"
    size         = "0"
    mountpoint   = "/"
  }
  image_chroot_env             = ["PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin"]
  qemu_binary_destination_path = "/usr/bin/qemu-arm-static"
  qemu_binary_source_path      = "/usr/bin/qemu-arm-static"
}

build {
  sources = ["source.arm.rpi"]

  provisioner "file" {
    sources     = ["target/librespot", "amp-httpd/target/amp-httpd-rpi", "target/nqptp", "target/shairport-sync"]
    destination = "/usr/local/bin/"
  }

  provisioner "file" {
    source      = "/root/.ssh/id_rsa.pub"
    destination = "/tmp/"
  }

  provisioner "shell" {
    environment_vars = [
      "WIFI_SSID=${var.wifi_ssid}",
      "WIFI_PASSWORD=${var.wifi_password}",
      "TZ_FILE=${var.tz}"
    ]
    scripts = [
      "scripts/system",
      "scripts/ssh",
      "scripts/net",
      "scripts/sound",
      "scripts/amp",
      "scripts/spotify",
      "scripts/shairport"
    ]
  }

  post-processor "shell-local" {
    inline = ["cp /opt/build.img /build/target/rpi.img"]
  }
}
