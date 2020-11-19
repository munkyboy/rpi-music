variable "wifi_ssid" {
  type    = string
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

source "arm" "rpi" {
  file_checksum         = "4522df4a29f9aac4b0166fbfee9f599dab55a997c855702bfe35329c13334668"
  file_checksum_type    = "sha256"
  file_target_extension = "zip"
  file_urls             = ["https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2020-08-24/2020-08-20-raspios-buster-armhf-lite.zip"]
  image_build_method    = "reuse"
  image_chroot_env      = ["PATH=/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin"]
  image_partitions {
    filesystem   = "vfat"
    mountpoint   = "/boot"
    name         = "boot"
    size         = "256M"
    start_sector = "8192"
    type         = "c"
  }
  image_partitions {
    filesystem   = "ext4"
    mountpoint   = "/"
    name         = "root"
    size         = "0"
    start_sector = "532480"
    type         = "83"
  }
  image_path                   = "/opt/build.img"
  image_size                   = "2G"
  image_type                   = "dos"
  qemu_binary_destination_path = "/usr/bin/qemu-arm-static"
  qemu_binary_source_path      = "/usr/bin/qemu-arm-static"
}

build {
  sources = ["source.arm.rpi"]

  provisioner "file" {
    sources     = ["target/librespot"]
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
      "scripts/spotify"
    ]
  }

  post-processor "shell-local" {
    inline = ["cp /opt/build.img /build/target/rpi.img"]
  }
}
