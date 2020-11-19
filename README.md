# Raspberry Pi Music Server

This repo builds a rapbian image which includes;

* librespot (Spotify connect service)
* controller for Monoprice 6 zone amp

The amp is expected to be plugged into a [controllable power
strip](https://dlidirect.com/products/iot-power-relay). The power strip is then
connected to the raspberry pi via a GPIO pin.

## Pre-requisites

* docker
* make

## Building

Create a file named `settings.auto.pkrvars.hcl` in the root of this repo. It
should contain packer variable declarations. e.g.

```
wifi_ssid = "home"
wifi_password = "secret"
```

See the variable stanzas in `./rpi.pkr.hcl` for the complete list.

To build, run:

```
make
```

## Testing

To boot the image locally for testing, run:

```
make run
```

## Deploying

Attach a SD card to your host computer then run:

```
make copy
```

You will be prompted to confirm the target disk.
